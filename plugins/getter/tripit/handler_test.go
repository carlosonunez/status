package tripit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/carlosonunez/status/pkg/pluginsdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newHandlerWithServer returns a Handler wired to a mock TripIt server.
func newHandlerWithServer(srv *httptest.Server) *Handler {
	h := New()
	h.newClient = func(_, _, _, _ string) *Client {
		return newClientWithBase(srv.URL)
	}
	h.now = func() time.Time {
		// Freeze time inside the trip: 2019-12-02 06:55 UTC
		// AA356 departs 2019-12-01T17:11-06:00 (= 2019-12-01T23:11 UTC)
		// AA2360 arrives 2019-12-05T17:58-06:00 (= 2019-12-05T23:58 UTC)
		return time.Date(2019, 12, 2, 6, 55, 0, 0, time.UTC)
	}
	return h
}

// buildTripServer returns an httptest.Server that returns the given response body.
func buildTripServer(t *testing.T, body any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(body))
	}))
}

func TestHandlerMetadata(t *testing.T) {
	h := New()
	m := h.Metadata()
	assert.Equal(t, "tripit", m.Name)
	assert.Equal(t, "5m", m.MinInterval)
	assert.True(t, m.SupportsAuth)
	assert.NotEmpty(t, m.ParamSpecs)
}

type handlerGetEventsTest struct {
	TestName      string
	APIResponse   any
	Params        map[string]any
	NowFunc       func() time.Time
	WantEventLen  int
	WantFirstCity string
	WantFlights   int
	WantError     bool
}

func (tc handlerGetEventsTest) RunTest(t *testing.T) {
	t.Helper()
	srv := buildTripServer(t, tc.APIResponse)
	defer srv.Close()

	h := newHandlerWithServer(srv)
	if tc.NowFunc != nil {
		h.now = tc.NowFunc
	}

	events, err := h.GetEvents(tc.Params)
	if tc.WantError {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)
	assert.Len(t, events, tc.WantEventLen)
	if tc.WantEventLen > 0 && tc.WantFirstCity != "" {
		assert.Equal(t, tc.WantFirstCity, events[0].Meta["city"])
	}
	if tc.WantEventLen > 0 && tc.WantFlights >= 0 {
		flights, ok := events[0].Meta["flights"].([]map[string]any)
		if ok {
			assert.Len(t, flights, tc.WantFlights)
		}
	}
}

func TestHandlerGetEvents(t *testing.T) {
	// Trip with two flights — frozen now is inside the trip window.
	activeTripWithFlights := listTripResponse{
		Trip: oneOrMany[trip]{
			{
				ID:              "293554134",
				DisplayName:     "Work: Test Client - Week 2",
				PrimaryLocation: "Omaha, NE",
				StartDate:       "2019-12-01",
				EndDate:         "2019-12-05",
				RelativeURL:     "/trip/show/id/293554134",
			},
		},
		AirObject: oneOrMany[airObject]{
			{
				TripID: "293554134",
				Segment: oneOrMany[segment]{
					{
						MarketingAirlineCode: "AA", MarketingFlightNumber: "356",
						StartAirportCode: "DFW", EndAirportCode: "OMA",
						StartDateTime: dateTime{Date: "2019-12-01", Time: "17:11:00", UTCOffset: "-06:00"},
						EndDateTime:   dateTime{Date: "2019-12-01", Time: "18:56:00", UTCOffset: "-06:00"},
					},
					{
						MarketingAirlineCode: "AA", MarketingFlightNumber: "2360",
						StartAirportCode: "OMA", EndAirportCode: "DFW",
						StartDateTime: dateTime{Date: "2019-12-05", Time: "16:02:00", UTCOffset: "-06:00"},
						EndDateTime:   dateTime{Date: "2019-12-05", Time: "17:58:00", UTCOffset: "-06:00"},
					},
				},
			},
		},
	}

	tripWithoutFlights := listTripResponse{
		Trip: oneOrMany[trip]{
			{
				ID:              "123456789",
				DisplayName:     "Personal: Some Trip",
				PrimaryLocation: "Dayton, OH",
				StartDate:       "2019-12-15",
				EndDate:         "2019-12-19",
				RelativeURL:     "/trip/show/id/123456789",
			},
		},
	}

	// "now" inside the personal trip (no flights; uses start/end date).
	insideNoFlightTrip := func() time.Time {
		return time.Date(2019, 12, 17, 12, 0, 0, 0, time.UTC)
	}
	// "now" before any trip starts.
	beforeTrip := func() time.Time {
		return time.Date(2019, 11, 30, 9, 0, 0, 0, time.UTC)
	}
	// "now" after trip ends.
	afterTrip := func() time.Time {
		return time.Date(2019, 12, 20, 0, 0, 0, 0, time.UTC)
	}

	tests := []handlerGetEventsTest{
		{
			TestName:    "no_trips",
			APIResponse: map[string]any{"timestamp": "1234"},
			NowFunc:     beforeTrip,
			WantEventLen: 0,
		},
		{
			TestName:      "active_trip_with_flights",
			APIResponse:   activeTripWithFlights,
			WantEventLen:  1,
			WantFirstCity: "Omaha, NE",
		},
		{
			TestName:     "now_before_starts_on",
			APIResponse:  activeTripWithFlights,
			NowFunc:      beforeTrip,
			WantEventLen: 0,
		},
		{
			TestName:     "now_after_ends_on",
			APIResponse:  activeTripWithFlights,
			NowFunc:      afterTrip,
			WantEventLen: 0,
		},
		{
			TestName:      "trip_without_flights_active",
			APIResponse:   tripWithoutFlights,
			NowFunc:       insideNoFlightTrip,
			WantEventLen:  1,
			WantFirstCity: "Dayton, OH",
		},
		{
			TestName:     "trip_without_flights_before",
			APIResponse:  tripWithoutFlights,
			NowFunc:      beforeTrip,
			WantEventLen: 0,
		},
		{
			TestName:  "api_error_propagated",
			WantError: true,
			// No server set up; handler uses a server that returns 500.
		},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			if tc.TestName == "api_error_propagated" {
				// Use a server that always returns 500.
				errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
				defer errSrv.Close()
				h := newHandlerWithServer(errSrv)
				_, err := h.GetEvents(map[string]any{
					"consumer_key": "k", "consumer_secret": "s",
					"access_token": "t", "token_secret": "ts",
				})
				require.Error(t, err)
				return
			}
			tc.RunTest(t)
		})
	}
}

func TestHandlerTripEndedNote(t *testing.T) {
	body := listTripResponse{
		Trip: oneOrMany[trip]{
			{
				ID:              "345678901",
				DisplayName:     "Personal: Some Trip That Ended",
				PrimaryLocation: "Dallas, TX",
				StartDate:       "2019-11-27",
				EndDate:         "2019-12-01",
				RelativeURL:     "/trip/show/id/345678901",
			},
		},
		NoteObject: oneOrMany[note]{
			{TripID: "345678901", DisplayName: "TRIP_ENDED"},
		},
	}
	srv := buildTripServer(t, body)
	defer srv.Close()

	h := newHandlerWithServer(srv)
	// Now is inside the window but TRIP_ENDED note is present.
	h.now = func() time.Time {
		return time.Date(2019, 11, 28, 12, 0, 0, 0, time.UTC)
	}

	events, err := h.GetEvents(map[string]any{
		"consumer_key": "k", "consumer_secret": "s",
		"access_token": "t", "token_secret": "ts",
	})
	require.NoError(t, err)
	// Trip is in the window but marked ended → still returned, ended=true.
	require.Len(t, events, 1)
	assert.Equal(t, true, events[0].Meta["ended"])
}

func TestHandlerIngressFromParam(t *testing.T) {
	body := listTripResponse{
		Trip: oneOrMany[trip]{
			{
				ID: "293554134", DisplayName: "Work Trip",
				PrimaryLocation: "Omaha, NE",
				StartDate:       "2019-12-01", EndDate: "2019-12-05",
				RelativeURL: "/trip/show/id/293554134",
			},
		},
		AirObject: oneOrMany[airObject]{
			{
				TripID: "293554134",
				Segment: oneOrMany[segment]{
					{
						MarketingAirlineCode: "AA", MarketingFlightNumber: "356",
						StartAirportCode: "DFW", EndAirportCode: "OMA",
						StartDateTime: dateTime{Date: "2019-12-01", Time: "17:11:00", UTCOffset: "-06:00"},
						EndDateTime:   dateTime{Date: "2019-12-01", Time: "18:56:00", UTCOffset: "-06:00"},
					},
				},
			},
		},
	}
	srv := buildTripServer(t, body)
	defer srv.Close()

	h := newHandlerWithServer(srv)
	// Now is exactly at depart + 30min ingress.
	departUTC := time.Date(2019, 12, 1, 23, 11, 0, 0, time.UTC) // 17:11 CST = 23:11 UTC
	h.now = func() time.Time { return departUTC.Add(31 * time.Minute) }

	// With ingress_minutes=30, startsAt = depart+30m → now is just inside.
	events, err := h.GetEvents(map[string]any{
		"consumer_key": "k", "consumer_secret": "s",
		"access_token": "t", "token_secret": "ts",
		"ingress_minutes": 30,
	})
	require.NoError(t, err)
	assert.Len(t, events, 1)
}

func TestHandlerAuthenticateOutputsURL(t *testing.T) {
	requestTokenCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/request_token":
			requestTokenCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("oauth_token=myreqtoken&oauth_token_secret=myreqsecret"))
		case "/oauth/access_token":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("oauth_token=myaccesstoken&oauth_token_secret=myaccesssecret"))
		case "/callback":
			// Simulate TripIt redirecting back.
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	h := newHandlerWithServer(srv)
	// Override the authorization URL builder to point at our test server.
	h.SetAuthorizeURL(func(token string, callbackURL string) string {
		return srv.URL + "/authorize?oauth_token=" + token + "&oauth_callback=" + callbackURL
	})
	// Immediately simulate the callback arriving.
	h.SetAwaitCallback(func(_ <-chan CallbackResult) (CallbackResult, error) {
		return CallbackResult{token: "myreqtoken", verifier: "myverifier"}, nil
	})

	var out bytes.Buffer
	tokens, err := h.Authenticate(map[string]any{
		"consumer_key":    "ckey",
		"consumer_secret": "csecret",
	}, &out)

	require.NoError(t, err)
	assert.True(t, requestTokenCalled)
	assert.Contains(t, out.String(), "tripit.com/oauth/authorize")
	assert.Equal(t, "myaccesstoken", tokens["access_token"])
	assert.Equal(t, "myaccesssecret", tokens["token_secret"])
}

// Ensure Handler implements the pluginsdk interfaces at compile time.
var _ pluginsdk.GetterHandler = (*Handler)(nil)
var _ pluginsdk.AuthHandler = (*Handler)(nil)
