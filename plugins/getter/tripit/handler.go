package tripit

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/carlosonunez/status/pkg/pluginsdk"
	log "github.com/sirupsen/logrus"
)

const (
	tripItAuthorizeURL = "https://www.tripit.com/oauth/authorize"
)

// CallbackResult carries the OAuth token and verifier received from TripIt's
// callback redirect.
type CallbackResult struct {
	token    string
	verifier string
	err      error
}

// callbackResult is an alias so internal code can use the unexported name.
type callbackResult = CallbackResult

// NewCallbackResult constructs a CallbackResult. Used by integration tests
// that live in package tripit_test.
func NewCallbackResult(token, verifier string, err error) CallbackResult {
	return CallbackResult{token: token, verifier: verifier, err: err}
}

// Handler implements pluginsdk.GetterHandler and pluginsdk.AuthHandler for
// the TripIt event getter.
type Handler struct {
	// newClient is a factory for creating API clients; injectable for tests.
	newClient func(consumerKey, consumerSecret, accessToken, tokenSecret string) *Client
	// now returns the current time; injectable for tests.
	now func() time.Time
	// authorizeURL builds the TripIt authorization URL; injectable for tests.
	authorizeURL func(token, callbackURL string) string
	// awaitCallback waits for the OAuth callback; injectable for tests.
	awaitCallback func(ch <-chan CallbackResult) (CallbackResult, error)
}

// New returns a new Handler with production defaults.
func New() *Handler {
	return &Handler{
		newClient: NewClient,
		now:       time.Now,
		authorizeURL: func(token, callbackURL string) string {
			return fmt.Sprintf("%s?oauth_token=%s&oauth_callback=%s",
				tripItAuthorizeURL, token, callbackURL)
		},
		awaitCallback: func(ch <-chan CallbackResult) (CallbackResult, error) {
			res := <-ch
			return res, res.err
		},
	}
}

// SetAwaitCallback replaces the callback-wait function. Used in tests to
// simulate an immediate callback without starting a real browser.
func (h *Handler) SetAwaitCallback(fn func(<-chan CallbackResult) (CallbackResult, error)) {
	h.awaitCallback = fn
}

// SetAuthorizeURL replaces the auth URL builder. Used in integration tests to
// intercept the URL before it is printed.
func (h *Handler) SetAuthorizeURL(fn func(token, callbackURL string) string) {
	h.authorizeURL = fn
}

// AuthorizeURLFunc returns the current auth URL builder. Used in tests to
// wrap the default builder.
func (h *Handler) AuthorizeURLFunc() func(token, callbackURL string) string {
	return h.authorizeURL
}

// Metadata returns the plugin's self-description.
func (h *Handler) Metadata() pluginsdk.GetterMetadata {
	return pluginsdk.GetterMetadata{
		Name:         "tripit",
		MinInterval:  "5m",
		SupportsAuth: true,
		ParamSpecs: []pluginsdk.ParamSpec{
			{Name: "consumer_key", Description: "TripIt OAuth 1.0 consumer key", Required: true, Type: "string"},
			{Name: "consumer_secret", Description: "TripIt OAuth 1.0 consumer secret", Required: true, Type: "string"},
			{Name: "access_token", Description: "OAuth access token (from token store)", Required: true, Type: "string"},
			{Name: "token_secret", Description: "OAuth token secret (from token store)", Required: true, Type: "string"},
			{Name: "ingress_minutes", Description: "Minutes added to first-flight departure (overrides TRIPIT_INGRESS_TIME_MINUTES)", Required: false, Type: "int"},
		},
	}
}

// GetEvents fetches currently-active TripIt trips and returns them as Events.
func (h *Handler) GetEvents(params map[string]any) ([]pluginsdk.Event, error) {
	cKey, _ := params["consumer_key"].(string)
	cSecret, _ := params["consumer_secret"].(string)
	aToken, _ := params["access_token"].(string)
	tSecret, _ := params["token_secret"].(string)

	var ingressOverride *int
	if v, ok := params["ingress_minutes"]; ok {
		switch val := v.(type) {
		case int:
			ingressOverride = &val
		case float64:
			i := int(val)
			ingressOverride = &i
		}
	}
	ingress := ingressSeconds(ingressOverride)

	client := h.newClient(cKey, cSecret, aToken, tSecret)
	resp, err := client.ListTrips(context.Background())
	if err != nil {
		return nil, fmt.Errorf("tripit: list trips: %w", err)
	}

	now := h.now()
	var events []pluginsdk.Event

	trips := []trip(resp.Trip)
	airObjects := []airObject(resp.AirObject)
	notes := []note(resp.NoteObject)

	for _, t := range trips {
		flights, err := resolveFlights(airObjects, t.ID)
		if err != nil {
			log.WithError(err).Warnf("tripit: skipping trip %s: could not resolve flights", t.ID)
			continue
		}

		startsAt := resolveStartTime(t, flights, ingress)
		endsAt := resolveEndTime(t, flights)
		ended := tripEnded(endsAt, notes, t.ID, now)

		// Only return trips that are currently active.
		if now.Before(startsAt) || now.After(endsAt) {
			log.Debugf("tripit: trip %q not active at %s (starts %s ends %s)", t.DisplayName, now, startsAt, endsAt)
			continue
		}

		flightMeta := buildFlightMeta(flights)
		events = append(events, pluginsdk.Event{
			Calendar: "tripit",
			Title:    t.DisplayName,
			StartsAt: startsAt.UTC(),
			EndsAt:   endsAt.UTC(),
			AllDay:   false,
			Meta: map[string]any{
				"city":    resolveLocation(t),
				"link":    "https://www.tripit.com" + t.RelativeURL,
				"ended":   ended,
				"flights": flightMeta,
			},
		})
	}
	return events, nil
}

// Authenticate runs the TripIt OAuth 1.0 dance:
//  1. Start a local callback server on a random port.
//  2. Obtain a request token.
//  3. Print the authorization URL to out.
//  4. Wait for the callback with the verifier.
//  5. Exchange for an access token.
//  6. Return {"access_token": "…", "token_secret": "…"}.
func (h *Handler) Authenticate(params map[string]any, out io.Writer) (map[string]string, error) {
	cKey, _ := params["consumer_key"].(string)
	cSecret, _ := params["consumer_secret"].(string)

	// Start a callback listener on a random free port.
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("tripit: start callback server: %w", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://localhost:%d/callback", port)

	client := h.newClient(cKey, cSecret, "", "")
	reqToken, reqSecret, err := client.GetRequestToken(context.Background(), cKey, cSecret, callbackURL)
	if err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("tripit: get request token: %w", err)
	}

	authURL := h.authorizeURL(reqToken, callbackURL)
	fmt.Fprintf(out, "\nOpen this URL to authorise status with TripIt:\n\n  %s\n\nWaiting for authorisation...\n", authURL)

	// Serve the callback in the background.
	callbackCh := make(chan CallbackResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		callbackCh <- CallbackResult{
			token:    q.Get("oauth_token"),
			verifier: q.Get("oauth_verifier"),
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()

	cbResult, err := h.awaitCallback(callbackCh)
	if err != nil {
		return nil, fmt.Errorf("tripit: await callback: %w", err)
	}

	accessToken, tokenSecret, err := client.GetAccessToken(
		context.Background(), cKey, cSecret, cbResult.token, reqSecret, cbResult.verifier)
	if err != nil {
		return nil, fmt.Errorf("tripit: exchange access token: %w", err)
	}

	fmt.Fprintln(out, "\nAuthorised successfully.")
	return map[string]string{
		"access_token": accessToken,
		"token_secret": tokenSecret,
	}, nil
}

// buildFlightMeta converts resolved flights to the Meta map format.
func buildFlightMeta(flights []flight) []map[string]any {
	out := make([]map[string]any, 0, len(flights))
	for _, f := range flights {
		out = append(out, map[string]any{
			"flight_number": f.FlightNumber,
			"origin":        f.Origin,
			"destination":   f.Destination,
			"depart_time":   f.DepartTime.UTC().Format(time.RFC3339),
			"arrive_time":   f.ArriveTime.UTC().Format(time.RFC3339),
			"offset":        f.Offset,
		})
	}
	return out
}
