package tripit

import "encoding/json"

// listTripResponse is the top-level response from GET /v1/list/trip.
// TripIt may return Trip, AirObject and NoteObject as either a single JSON
// object or a JSON array; oneOrMany handles both cases transparently.
type listTripResponse struct {
	Trip       oneOrMany[trip]      `json:"Trip"`
	AirObject  oneOrMany[airObject] `json:"AirObject"`
	NoteObject oneOrMany[note]      `json:"NoteObject"`
}

type trip struct {
	ID              string `json:"id"`
	DisplayName     string `json:"display_name"`
	PrimaryLocation string `json:"primary_location"`
	StartDate       string `json:"start_date"` // YYYY-MM-DD
	EndDate         string `json:"end_date"`   // YYYY-MM-DD
	RelativeURL     string `json:"relative_url"`
}

type airObject struct {
	TripID  string             `json:"trip_id"`
	Segment oneOrMany[segment] `json:"Segment"`
}

type segment struct {
	MarketingAirlineCode  string   `json:"marketing_airline_code"`
	MarketingFlightNumber string   `json:"marketing_flight_number"`
	StartAirportCode      string   `json:"start_airport_code"`
	EndAirportCode        string   `json:"end_airport_code"`
	StartDateTime         dateTime `json:"StartDateTime"`
	EndDateTime           dateTime `json:"EndDateTime"`
}

type dateTime struct {
	Date      string `json:"date"`       // YYYY-MM-DD
	Time      string `json:"time"`       // HH:MM:SS
	UTCOffset string `json:"utc_offset"` // e.g. "-06:00"
}

type note struct {
	TripID      string `json:"trip_id"`
	DisplayName string `json:"display_name"`
}

// oneOrMany[T] unmarshals a JSON value that may be either a single T or a
// []T. TripIt uses both forms depending on how many objects are present.
type oneOrMany[T any] []T

func (o *oneOrMany[T]) UnmarshalJSON(data []byte) error {
	// Try array first.
	var many []T
	if err := json.Unmarshal(data, &many); err == nil {
		*o = many
		return nil
	}
	// Fall back to single object.
	var one T
	if err := json.Unmarshal(data, &one); err != nil {
		return err
	}
	*o = []T{one}
	return nil
}
