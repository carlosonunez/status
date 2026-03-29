package tripit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseTime is a test helper that parses an RFC3339 string or fails the test.
func parseTime(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return ts
}

// ── normaliseFlightTimeToTZ ───────────────────────────────────────────────────

type normaliseFlightTimeToTZTest struct {
	TestName  string
	DT        dateTime
	Want      time.Time
	WantError bool
}

func (tc normaliseFlightTimeToTZTest) RunTest(t *testing.T) {
	t.Helper()
	got, err := normaliseFlightTimeToTZ(tc.DT)
	if tc.WantError {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)
	assert.Equal(t, tc.Want, got)
}

func TestNormaliseFlightTimeToTZ(t *testing.T) {
	tests := []normaliseFlightTimeToTZTest{
		{
			TestName: "chicago_minus_six",
			DT:       dateTime{Date: "2019-12-01", Time: "17:11:00", UTCOffset: "-06:00"},
			Want:     parseTime(t, "2019-12-01T17:11:00-06:00"),
		},
		{
			TestName: "eastern_minus_five",
			DT:       dateTime{Date: "2019-11-27", Time: "11:19:00", UTCOffset: "-05:00"},
			Want:     parseTime(t, "2019-11-27T11:19:00-05:00"),
		},
		{
			TestName: "pacific_minus_eight",
			DT:       dateTime{Date: "2019-11-27", Time: "13:01:00", UTCOffset: "-08:00"},
			Want:     parseTime(t, "2019-11-27T13:01:00-08:00"),
		},
		{
			TestName:  "bad_date",
			DT:        dateTime{Date: "not-a-date", Time: "00:00:00", UTCOffset: "-06:00"},
			WantError: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) { tc.RunTest(t) })
	}
}

// ── resolveFlights ────────────────────────────────────────────────────────────

type resolveFlightsTest struct {
	TestName   string
	AirObjects []airObject
	TripID     string
	WantLen    int
	WantFirst  *flight // nil = don't check
	WantError  bool
}

func (tc resolveFlightsTest) RunTest(t *testing.T) {
	t.Helper()
	got, err := resolveFlights(tc.AirObjects, tc.TripID)
	if tc.WantError {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)
	assert.Len(t, got, tc.WantLen)
	if tc.WantFirst != nil && len(got) > 0 {
		assert.Equal(t, tc.WantFirst.FlightNumber, got[0].FlightNumber)
		assert.Equal(t, tc.WantFirst.Origin, got[0].Origin)
		assert.Equal(t, tc.WantFirst.Destination, got[0].Destination)
	}
}

func TestResolveFlights(t *testing.T) {
	twoSegmentAir := airObject{
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
	}
	singleSegmentAir := airObject{
		TripID: "123",
		Segment: oneOrMany[segment]{
			{
				MarketingAirlineCode: "AA", MarketingFlightNumber: "1",
				StartAirportCode: "JFK", EndAirportCode: "LAX",
				StartDateTime: dateTime{Date: "2019-11-27", Time: "11:19:00", UTCOffset: "-05:00"},
				EndDateTime:   dateTime{Date: "2019-11-27", Time: "13:01:00", UTCOffset: "-08:00"},
			},
		},
	}

	tests := []resolveFlightsTest{
		{
			TestName:   "no_air_objects",
			AirObjects: nil,
			TripID:     "293554134",
			WantLen:    0,
		},
		{
			TestName:   "air_object_for_different_trip",
			AirObjects: []airObject{twoSegmentAir},
			TripID:     "999",
			WantLen:    0,
		},
		{
			TestName:   "two_segments_sorted_by_depart_time",
			AirObjects: []airObject{twoSegmentAir},
			TripID:     "293554134",
			WantLen:    2,
			WantFirst:  &flight{FlightNumber: "AA356", Origin: "DFW", Destination: "OMA"},
		},
		{
			TestName:   "single_segment",
			AirObjects: []airObject{singleSegmentAir},
			TripID:     "123",
			WantLen:    1,
			WantFirst:  &flight{FlightNumber: "AA1", Origin: "JFK", Destination: "LAX"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) { tc.RunTest(t) })
	}
}

// ── resolveStartTime ──────────────────────────────────────────────────────────

type resolveStartTimeTest struct {
	TestName   string
	Trip       trip
	Flights    []flight
	IngressSec int
	Want       time.Time
}

func (tc resolveStartTimeTest) RunTest(t *testing.T) {
	t.Helper()
	got := resolveStartTime(tc.Trip, tc.Flights, tc.IngressSec)
	assert.Equal(t, tc.Want, got)
}

func TestResolveStartTime(t *testing.T) {
	depart := parseTime(t, "2019-12-01T17:11:00-06:00")
	tests := []resolveStartTimeTest{
		{
			TestName:   "no_flights_uses_start_date",
			Trip:       trip{ID: "1", StartDate: "2019-12-15"},
			Flights:    nil,
			IngressSec: 0,
			Want:       mustParseDate("2019-12-15"),
		},
		{
			TestName:   "first_flight_depart_no_ingress",
			Trip:       trip{ID: "1", StartDate: "2019-12-01"},
			Flights:    []flight{{DepartTime: depart}},
			IngressSec: 0,
			Want:       depart,
		},
		{
			TestName:   "first_flight_depart_plus_ingress",
			Trip:       trip{ID: "1", StartDate: "2019-12-01"},
			Flights:    []flight{{DepartTime: depart}},
			IngressSec: 90 * 60,
			Want:       depart.Add(90 * time.Minute),
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) { tc.RunTest(t) })
	}
}

// ── resolveEndTime ────────────────────────────────────────────────────────────

type resolveEndTimeTest struct {
	TestName string
	Trip     trip
	Flights  []flight
	Want     time.Time
}

func (tc resolveEndTimeTest) RunTest(t *testing.T) {
	t.Helper()
	got := resolveEndTime(tc.Trip, tc.Flights)
	assert.Equal(t, tc.Want, got)
}

func TestResolveEndTime(t *testing.T) {
	arrive := parseTime(t, "2019-12-05T17:58:00-06:00")
	tests := []resolveEndTimeTest{
		{
			TestName: "no_flights_uses_end_date",
			Trip:     trip{ID: "1", EndDate: "2019-12-19"},
			Flights:  nil,
			Want:     mustParseDate("2019-12-19"),
		},
		{
			TestName: "last_flight_arrive_time",
			Trip:     trip{ID: "1", EndDate: "2019-12-05"},
			Flights: []flight{
				{ArriveTime: parseTime(t, "2019-12-01T18:56:00-06:00")},
				{ArriveTime: arrive},
			},
			Want: arrive,
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) { tc.RunTest(t) })
	}
}

// ── resolveLocation ───────────────────────────────────────────────────────────

type resolveLocationTest struct {
	TestName string
	Trip     trip
	Want     string
}

func (tc resolveLocationTest) RunTest(t *testing.T) {
	t.Helper()
	assert.Equal(t, tc.Want, resolveLocation(tc.Trip))
}

func TestResolveLocation(t *testing.T) {
	tests := []resolveLocationTest{
		{
			TestName: "has_primary_location",
			Trip:     trip{ID: "1", PrimaryLocation: "Omaha, NE"},
			Want:     "Omaha, NE",
		},
		{
			TestName: "missing_primary_location",
			Trip:     trip{ID: "1"},
			Want:     "Anywhere, Earth",
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) { tc.RunTest(t) })
	}
}

// ── tripEnded ────────────────────────────────────────────────────────────────

type tripEndedTest struct {
	TestName string
	EndsAt   time.Time
	Notes    []note
	TripID   string
	Now      time.Time
	Want     bool
}

func (tc tripEndedTest) RunTest(t *testing.T) {
	t.Helper()
	assert.Equal(t, tc.Want, tripEnded(tc.EndsAt, tc.Notes, tc.TripID, tc.Now))
}

func TestTripEnded(t *testing.T) {
	endsAt := parseTime(t, "2019-12-05T17:58:00-06:00")
	tests := []tripEndedTest{
		{
			TestName: "not_ended_now_before_end",
			EndsAt:   endsAt,
			TripID:   "1",
			Now:      endsAt.Add(-1 * time.Hour),
			Want:     false,
		},
		{
			TestName: "ended_now_after_end",
			EndsAt:   endsAt,
			TripID:   "1",
			Now:      endsAt.Add(1 * time.Hour),
			Want:     true,
		},
		{
			TestName: "ended_by_trip_ended_note",
			EndsAt:   endsAt,
			TripID:   "345678901",
			Notes:    []note{{TripID: "345678901", DisplayName: "TRIP_ENDED"}},
			Now:      endsAt.Add(-1 * time.Hour),
			Want:     true,
		},
		{
			TestName: "trip_ended_note_for_different_trip",
			EndsAt:   endsAt,
			TripID:   "1",
			Notes:    []note{{TripID: "345678901", DisplayName: "TRIP_ENDED"}},
			Now:      endsAt.Add(-1 * time.Hour),
			Want:     false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) { tc.RunTest(t) })
	}
}

// ── ingressSeconds ────────────────────────────────────────────────────────────

func TestIngressSeconds(t *testing.T) {
	t.Run("default_is_zero", func(t *testing.T) {
		t.Setenv("TRIPIT_INGRESS_TIME_MINUTES", "")
		assert.Equal(t, 0, ingressSeconds(nil))
	})
	t.Run("env_var_used_when_no_param", func(t *testing.T) {
		t.Setenv("TRIPIT_INGRESS_TIME_MINUTES", "90")
		assert.Equal(t, 90*60, ingressSeconds(nil))
	})
	t.Run("param_overrides_env_var", func(t *testing.T) {
		t.Setenv("TRIPIT_INGRESS_TIME_MINUTES", "90")
		v := 30
		assert.Equal(t, 30*60, ingressSeconds(&v))
	})
}

// ── helpers ───────────────────────────────────────────────────────────────────

// mustParseDate parses a YYYY-MM-DD date as midnight UTC.
func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}
