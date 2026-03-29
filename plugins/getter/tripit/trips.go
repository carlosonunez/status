package tripit

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

// flight is the resolved internal representation of a single flight leg.
type flight struct {
	FlightNumber string
	Origin       string
	Destination  string
	DepartTime   time.Time
	ArriveTime   time.Time
	Offset       string
}

// normaliseFlightTimeToTZ converts a TripIt dateTime (date + time + utc_offset)
// into a time.Time with the offset preserved.
func normaliseFlightTimeToTZ(dt dateTime) (time.Time, error) {
	s := fmt.Sprintf("%sT%s%s", dt.Date, dt.Time, dt.UTCOffset)
	t, err := time.Parse("2006-01-02T15:04:05-07:00", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("tripit: parse datetime %q: %w", s, err)
	}
	return t, nil
}

// resolveFlights returns all flight segments belonging to tripID, sorted by
// departure time ascending.
func resolveFlights(airObjects []airObject, tripID string) ([]flight, error) {
	var flights []flight
	for _, ao := range airObjects {
		if ao.TripID != tripID {
			continue
		}
		for _, seg := range ao.Segment {
			depart, err := normaliseFlightTimeToTZ(seg.StartDateTime)
			if err != nil {
				return nil, fmt.Errorf("tripit: segment depart time: %w", err)
			}
			arrive, err := normaliseFlightTimeToTZ(seg.EndDateTime)
			if err != nil {
				return nil, fmt.Errorf("tripit: segment arrive time: %w", err)
			}
			flights = append(flights, flight{
				FlightNumber: seg.MarketingAirlineCode + seg.MarketingFlightNumber,
				Origin:       seg.StartAirportCode,
				Destination:  seg.EndAirportCode,
				DepartTime:   depart,
				ArriveTime:   arrive,
				Offset:       seg.StartDateTime.UTCOffset,
			})
		}
	}
	sort.Slice(flights, func(i, j int) bool {
		return flights[i].DepartTime.Before(flights[j].DepartTime)
	})
	return flights, nil
}

// resolveStartTime returns the trip start time.
// If flights exist, it is the first departure time plus ingressSec.
// Otherwise it falls back to the trip's start_date (midnight local).
func resolveStartTime(t trip, flights []flight, ingressSec int) time.Time {
	if len(flights) == 0 {
		return parseTripDate(t.StartDate)
	}
	return flights[0].DepartTime.Add(time.Duration(ingressSec) * time.Second)
}

// resolveEndTime returns the trip end time.
// If flights exist, it is the last arrival time.
// Otherwise it falls back to the trip's end_date (midnight local).
func resolveEndTime(t trip, flights []flight) time.Time {
	if len(flights) == 0 {
		return parseTripDate(t.EndDate)
	}
	return flights[len(flights)-1].ArriveTime
}

// resolveLocation returns the trip's primary location, or "Anywhere, Earth"
// when none is set.
func resolveLocation(t trip) string {
	if t.PrimaryLocation == "" {
		return "Anywhere, Earth"
	}
	return t.PrimaryLocation
}

// tripEnded returns true when the trip has ended — either because a
// TRIP_ENDED NoteObject is attached to the trip, or because now is after endsAt.
func tripEnded(endsAt time.Time, notes []note, tripID string, now time.Time) bool {
	for _, n := range notes {
		if n.TripID == tripID && n.DisplayName == "TRIP_ENDED" {
			return true
		}
	}
	return now.After(endsAt)
}

// ingressSeconds returns the ingress buffer in seconds.
// If minutesOverride is non-nil its value is used directly.
// Otherwise the TRIPIT_INGRESS_TIME_MINUTES env var is read.
// Defaults to 0 if neither is set.
func ingressSeconds(minutesOverride *int) int {
	if minutesOverride != nil {
		return *minutesOverride * 60
	}
	if raw := os.Getenv("TRIPIT_INGRESS_TIME_MINUTES"); raw != "" {
		if mins, err := strconv.Atoi(raw); err == nil {
			return mins * 60
		}
	}
	return 0
}

// parseTripDate parses a YYYY-MM-DD trip date as midnight in the local timezone,
// matching TripIt's convention for date-only fields.
func parseTripDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
