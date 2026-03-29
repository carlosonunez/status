package tripit

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

// Browser wraps chromedp to provide a Capybara-style API for interacting with
// the TripIt website during integration tests and the auth flow.
type Browser struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// Context returns the underlying chromedp context. Exposed for integration
// tests that need to run raw chromedp actions alongside the Browser helpers.
func (b *Browser) Context() context.Context { return b.ctx }

// NewBrowser starts a headless Chrome instance. If CHROME_REMOTE_URL is set,
// it connects to a remote Chrome over DevTools Protocol (useful in Docker CI).
func NewBrowser(ctx context.Context) (*Browser, error) {
	var allocCtx context.Context
	var cancel context.CancelFunc

	if remoteURL := os.Getenv("CHROME_REMOTE_URL"); remoteURL != "" {
		allocCtx, cancel = chromedp.NewRemoteAllocator(ctx, remoteURL)
	} else {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-default-browser-check", true),
		)
		allocCtx, cancel = chromedp.NewExecAllocator(ctx, opts...)
	}

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	combinedCancel := func() { browserCancel(); cancel() }

	if err := chromedp.Run(browserCtx); err != nil {
		combinedCancel()
		return nil, fmt.Errorf("tripit: start browser: %w", err)
	}
	return &Browser{ctx: browserCtx, cancel: combinedCancel}, nil
}

// SetGDPRCookies sets the TRUSTe consent cookies on tripit.com to suppress
// the GDPR popup during automated flows.
func (b *Browser) SetGDPRCookies() error {
	return chromedp.Run(b.ctx,
		chromedp.Navigate("https://www.tripit.com"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Run(ctx,
				chromedp.EvaluateAsDevTools(`document.cookie = "notice_gdpr_prefs=0,1,2:; domain=tripit.com"`, nil),
				chromedp.EvaluateAsDevTools(`document.cookie = "notice_preferences=2:; domain=tripit.com"`, nil),
			)
		}),
	)
}

// Visit navigates to the given URL.
func (b *Browser) Visit(url string) error {
	return chromedp.Run(b.ctx, chromedp.Navigate(url))
}

// FillIn sets the value of an input identified by its id attribute.
func (b *Browser) FillIn(id, value string) error {
	return chromedp.Run(b.ctx,
		chromedp.WaitVisible(`#`+id, chromedp.ByID),
		chromedp.SetValue(`#`+id, value, chromedp.ByID),
	)
}

// ClickButton clicks the first button or submit input whose visible text or
// id matches label.
func (b *Browser) ClickButton(label string) error {
	// Try by id first, then by visible text.
	sel := fmt.Sprintf(`//button[contains(text(),%q)] | //input[@type='submit'][@value=%q] | //button[@id=%q]`,
		label, label, label)
	return chromedp.Run(b.ctx,
		chromedp.WaitVisible(sel, chromedp.BySearch),
		chromedp.Click(sel, chromedp.BySearch),
	)
}

// PageSource returns the current page's source HTML.
func (b *Browser) PageSource() (string, error) {
	var src string
	err := chromedp.Run(b.ctx, chromedp.OuterHTML("html", &src))
	return src, err
}

// CurrentURL returns the browser's current URL.
func (b *Browser) CurrentURL() (string, error) {
	var u string
	err := chromedp.Run(b.ctx, chromedp.Location(&u))
	return u, err
}

// WaitForURL blocks until the browser navigates to a URL matching the given
// prefix, or until timeout elapses.
func (b *Browser) WaitForURL(prefix string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		u, err := b.CurrentURL()
		if err == nil {
			if len(u) >= len(prefix) && u[:len(prefix)] == prefix {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("tripit: timed out waiting for URL with prefix %q", prefix)
}

// Close shuts down the browser. Suppressed when KEEP_BROWSER_OPEN=true.
func (b *Browser) Close() {
	if os.Getenv("KEEP_BROWSER_OPEN") == "true" {
		return
	}
	b.cancel()
}
