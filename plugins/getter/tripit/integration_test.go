//go:build integration

package tripit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	tripit "github.com/carlosonunez/status/plugins/getter/tripit"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requiredEnv returns the value of an env var or skips the test if absent.
func requiredEnv(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("skipping integration test: %s not set", key)
	}
	return v
}

// newIntegrationBrowser creates a Browser for use in integration tests and
// registers cleanup.
func newIntegrationBrowser(t *testing.T) *tripit.Browser {
	t.Helper()
	b, err := tripit.NewBrowser(context.Background())
	require.NoError(t, err)
	require.NoError(t, b.SetGDPRCookies())
	t.Cleanup(b.Close)
	return b
}

// signIn navigates to the TripIt login page and authenticates.
func signIn(t *testing.T, b *tripit.Browser, email, password string) {
	t.Helper()
	require.NoError(t, b.Visit("https://www.tripit.com/account/login"))
	require.NoError(t, b.FillIn("email_address", email))
	require.NoError(t, b.FillIn("password", password))
	require.NoError(t, b.ClickButton("Sign In"))
	// Wait for the password field to disappear (login complete).
	_ = chromedp.Run(b.Context(), chromedp.WaitNotPresent(`#password`, chromedp.ByID))
}

// authorizeApp drives the browser to the TripIt OAuth authorization page and
// returns the raw callback response body.
func authorizeApp(t *testing.T, b *tripit.Browser, authURL string) string {
	t.Helper()
	var pageSource string
	err := chromedp.Run(b.Context(),
		chromedp.Navigate(authURL),
		chromedp.WaitVisible(`body`, chromedp.BySearch),
		chromedp.OuterHTML("html", &pageSource),
	)
	require.NoError(t, err)
	pageSource = strings.NewReplacer(
		`<html><head></head><body><pre style="word-wrap: break-word; white-space: pre-wrap;">`, "",
		`</pre></body></html>`, "",
	).Replace(pageSource)
	return pageSource
}

// createTrip logs into TripIt and creates a test trip for today/tomorrow,
// returning the new trip's ID.
func createTrip(t *testing.T, b *tripit.Browser, email, password string) string {
	t.Helper()
	require.NoError(t, b.SetGDPRCookies())
	signIn(t, b, email, password)

	today := time.Now().Format("01/02/2006")
	tomorrow := time.Now().Add(24 * time.Hour).Format("01/02/2006")

	require.NoError(t, b.Visit("https://www.tripit.com/trip/create"))
	require.NoError(t, b.FillIn("place", "Dallas, TX"))
	require.NoError(t, b.FillIn("display_name", fmt.Sprintf("Test Trip %d", time.Now().Unix())))
	require.NoError(t, b.FillIn("start_date", today))
	require.NoError(t, b.FillIn("end_date", tomorrow))
	require.NoError(t, b.ClickButton("Add Trip"))

	var currentURL string
	err := chromedp.Run(b.Context(),
		chromedp.WaitNotPresent(`#display_name`, chromedp.ByID),
		chromedp.Location(&currentURL),
	)
	require.NoError(t, err)

	parts := strings.Split(currentURL, "/")
	tripID := parts[len(parts)-1]
	require.NotEmpty(t, tripID)
	return tripID
}

// deleteTrip removes a trip from TripIt via the browser.
func deleteTrip(t *testing.T, b *tripit.Browser, tripID string) {
	t.Helper()
	err := chromedp.Run(b.Context(),
		chromedp.Navigate(fmt.Sprintf(
			"https://www.tripit.com/trips/delete/%s?redirect_url=https://www.tripit.com/trips",
			tripID,
		)),
		chromedp.WaitVisible(`//button[contains(text(),'Delete')]`, chromedp.BySearch),
		chromedp.Click(`//button[contains(text(),'Delete')]`, chromedp.BySearch),
		chromedp.WaitVisible(`//button[contains(text(),'Confirm Delete')]`, chromedp.BySearch),
		chromedp.Click(`//button[contains(text(),'Confirm Delete')]`, chromedp.BySearch),
	)
	require.NoError(t, err)
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestIntegrationAuthURL(t *testing.T) {
	cKey := requiredEnv(t, "TRIPIT_APP_CLIENT_ID")
	cSecret := requiredEnv(t, "TRIPIT_APP_CLIENT_SECRET")

	h := tripit.New()
	var out bytes.Buffer
	// Intercept before the callback wait so the test doesn't hang.
	h.SetAwaitCallback(func(_ <-chan tripit.CallbackResult) (tripit.CallbackResult, error) {
		return tripit.NewCallbackResult("", "", fmt.Errorf("test: intentional early exit"))
	})

	_, _ = h.Authenticate(map[string]any{
		"consumer_key":    cKey,
		"consumer_secret": cSecret,
	}, &out)

	// The auth URL should have been printed to out before the hang.
	authURLPattern := regexp.MustCompile(
		`https://www\.tripit\.com/oauth/authorize\?oauth_token=[a-zA-Z0-9]+&oauth_callback=http://localhost:\d+/callback`,
	)
	assert.Regexp(t, authURLPattern, out.String(), "auth URL not found in output")
}

func TestIntegrationFullAuthFlow(t *testing.T) {
	cKey := requiredEnv(t, "TRIPIT_APP_CLIENT_ID")
	cSecret := requiredEnv(t, "TRIPIT_APP_CLIENT_SECRET")
	email := requiredEnv(t, "TRIPIT_SANDBOX_ACCOUNT_EMAIL")
	password := requiredEnv(t, "TRIPIT_SANDBOX_ACCOUNT_PASSWORD")

	h := tripit.New()

	// Intercept the auth URL, drive the browser to authorize, let the callback
	// complete naturally.
	var capturedAuthURL string
	origAuthorize := h.AuthorizeURLFunc()
	h.SetAuthorizeURL(func(token, callbackURL string) string {
		url := origAuthorize(token, callbackURL)
		capturedAuthURL = url + "&is_sign_in=1"
		return url
	})

	// Run Authenticate in a goroutine (it blocks on the callback).
	type result struct {
		tokens map[string]string
		err    error
	}
	resultCh := make(chan result, 1)
	go func() {
		tokens, err := h.Authenticate(map[string]any{
			"consumer_key":    cKey,
			"consumer_secret": cSecret,
		}, &bytes.Buffer{})
		resultCh <- result{tokens, err}
	}()

	// Wait briefly for the URL to be set, then drive the browser.
	deadline := time.Now().Add(15 * time.Second)
	for capturedAuthURL == "" && time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)
	}
	require.NotEmpty(t, capturedAuthURL, "auth URL was not set in time")

	browser := newTripitBrowser(t)
	defer browser.close(t)
	browser.gdprCookies(t)
	browser.signIn(t, email, password)
	rawResp := browser.authorizeApp(t, capturedAuthURL)

	var callbackResp struct {
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal([]byte(rawResp), &callbackResp))
	assert.Equal(t, "ok", callbackResp.Status)

	// Wait for the auth goroutine to finish.
	select {
	case res := <-resultCh:
		require.NoError(t, res.err)
		assert.NotEmpty(t, res.tokens["access_token"])
		assert.NotEmpty(t, res.tokens["token_secret"])
	case <-time.After(30 * time.Second):
		t.Fatal("timeout waiting for auth to complete")
	}
}

func TestIntegrationGetEvents(t *testing.T) {
	requiredEnv(t, "TRIPIT_APP_CLIENT_ID")
	requiredEnv(t, "TRIPIT_APP_CLIENT_SECRET")
	accessToken := requiredEnv(t, "TRIPIT_ACCESS_TOKEN")
	tokenSecret := requiredEnv(t, "TRIPIT_TOKEN_SECRET")

	h := tripit.New()
	events, err := h.GetEvents(map[string]any{
		"consumer_key":    os.Getenv("TRIPIT_APP_CLIENT_ID"),
		"consumer_secret": os.Getenv("TRIPIT_APP_CLIENT_SECRET"),
		"access_token":    accessToken,
		"token_secret":    tokenSecret,
	})
	require.NoError(t, err)
	// We can't assert specific trips (the account may have none active right now),
	// but we can assert the call succeeds and returns a valid slice.
	assert.NotNil(t, events)
}

func TestIntegrationSetAndGetCurrentTrip(t *testing.T) {
	cKey := requiredEnv(t, "TRIPIT_APP_CLIENT_ID")
	cSecret := requiredEnv(t, "TRIPIT_APP_CLIENT_SECRET")
	email := requiredEnv(t, "TRIPIT_SANDBOX_ACCOUNT_EMAIL")
	password := requiredEnv(t, "TRIPIT_SANDBOX_ACCOUNT_PASSWORD")
	accessToken := requiredEnv(t, "TRIPIT_ACCESS_TOKEN")
	tokenSecret := requiredEnv(t, "TRIPIT_TOKEN_SECRET")

	browser := newTripitBrowser(t)
	defer browser.close(t)

	tripID := browser.createTrip(t, email, password)
	t.Cleanup(func() { browser.deleteTrip(t, tripID) })

	h := tripit.New()
	events, err := h.GetEvents(map[string]any{
		"consumer_key":    cKey,
		"consumer_secret": cSecret,
		"access_token":    accessToken,
		"token_secret":    tokenSecret,
	})
	require.NoError(t, err)
	require.NotEmpty(t, events, "expected at least one active event after creating trip")

	// The trip we just created should be present.
	found := false
	for _, e := range events {
		if strings.Contains(e.Title, "Test Trip") {
			found = true
			assert.Equal(t, "Dallas, TX", e.Meta["city"])
		}
	}
	assert.True(t, found, "created test trip not found in GetEvents response")
}
