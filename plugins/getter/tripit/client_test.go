package tripit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientListTrips(t *testing.T) {
	type clientListTripsTest struct {
		TestName       string
		ResponseStatus int
		ResponseBody   any
		WantError      bool
		WantTripCount  int
		WantAirCount   int
	}

	singleTripBody := listTripResponse{
		Trip: oneOrMany[trip]{
			{ID: "1", DisplayName: "Test Trip", StartDate: "2019-12-01", EndDate: "2019-12-05"},
		},
	}
	twoTripsBody := listTripResponse{
		Trip: oneOrMany[trip]{
			{ID: "1", DisplayName: "Trip A"},
			{ID: "2", DisplayName: "Trip B"},
		},
		AirObject: oneOrMany[airObject]{
			{TripID: "1", Segment: oneOrMany[segment]{
				{MarketingAirlineCode: "AA", MarketingFlightNumber: "356",
					StartAirportCode: "DFW", EndAirportCode: "OMA",
					StartDateTime: dateTime{Date: "2019-12-01", Time: "17:11:00", UTCOffset: "-06:00"},
					EndDateTime:   dateTime{Date: "2019-12-01", Time: "18:56:00", UTCOffset: "-06:00"}},
			}},
		},
	}

	tests := []clientListTripsTest{
		{
			TestName:       "single_trip_no_flights",
			ResponseStatus: http.StatusOK,
			ResponseBody:   singleTripBody,
			WantTripCount:  1,
			WantAirCount:   0,
		},
		{
			TestName:       "two_trips_with_flights",
			ResponseStatus: http.StatusOK,
			ResponseBody:   twoTripsBody,
			WantTripCount:  2,
			WantAirCount:   1,
		},
		{
			TestName:       "non_200_returns_error",
			ResponseStatus: http.StatusUnauthorized,
			ResponseBody:   map[string]string{"error": "unauthorized"},
			WantError:      true,
		},
		{
			TestName:       "empty_response_body_returns_error",
			ResponseStatus: http.StatusOK,
			ResponseBody:   "not json",
			WantError:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/list/trip", r.URL.Path)
				assert.Equal(t, "true", r.URL.Query().Get("include_objects"))
				assert.Equal(t, "json", r.URL.Query().Get("format"))

				w.WriteHeader(tc.ResponseStatus)
				_ = json.NewEncoder(w).Encode(tc.ResponseBody)
			}))
			defer srv.Close()

			client := newClientWithBase(srv.URL)
			resp, err := client.ListTrips(context.Background())

			if tc.WantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, []trip(resp.Trip), tc.WantTripCount)
			assert.Len(t, []airObject(resp.AirObject), tc.WantAirCount)
		})
	}
}

func TestClientGetRequestToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/oauth/request_token", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("oauth_token=reqtoken&oauth_token_secret=reqsecret"))
	}))
	defer srv.Close()

	client := newClientWithBase(srv.URL)
	token, secret, err := client.GetRequestToken(context.Background(), "ckey", "csecret", "http://localhost:9999/callback")
	require.NoError(t, err)
	assert.Equal(t, "reqtoken", token)
	assert.Equal(t, "reqsecret", secret)
}

func TestClientGetAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/oauth/access_token", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("oauth_token=accesstoken&oauth_token_secret=accesssecret"))
	}))
	defer srv.Close()

	client := newClientWithBase(srv.URL)
	token, secret, err := client.GetAccessToken(context.Background(), "ckey", "csecret", "reqtoken", "reqsecret", "verifier123")
	require.NoError(t, err)
	assert.Equal(t, "accesstoken", token)
	assert.Equal(t, "accesssecret", secret)
}
