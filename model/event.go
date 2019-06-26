package model

import (
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	"net/http"
)

//EventRequest is JSON payload for the http post request
type EventRequest struct {
	Schema        *gojsonschema.Schema `json:"-"`
	UnixTimestamp int64                `json:"unix_timestamp"`
	Username      *string              `json:"username"`   //string pointer so jsonschema can enforce required args, otherwise default value "" renders required useless
	EventID       *string              `json:"event_uuid"` //same as above
	IPAddress     *string              `json:"ip_address"` //same as above
}

//EventRequestValidated is a container for the event request after its been validated, so you don't
//have to work with pointers to strings, etc.
type EventRequestValidated struct {
	UnixTimestamp int64
	Username      string
	EventID       string
	IPAddress     string
}

//Bind on EventRequest will run after unmarshalling, we can focus on things like schema validation.
func (e *EventRequest) Bind(r *http.Request) error {
	if e.Schema == nil {
		return errors.New("failed to validate schema properly")
	}

	result, err := e.Schema.Validate(gojsonschema.NewGoLoader(e))
	if err != nil {
		return err
	}

	if !result.Valid() {
		var response error
		for _, desc := range result.Errors() {
			response = multierror.Append(response, errors.New(desc.String()))
		}
		return response
	}
	return nil
}

//Geo holds location information
type Geo struct {
	Lat    float64 `db:"lat" json:"lat"`
	Lon    float64 `db:"lon" json:"lon"`
	Radius uint16  `db:"radius" json:"radius"`
}

//IPAccess holds geolocation information and other data about access events
type IPAccess struct {
	Geo
	Speed     float64 `json:"speed"`
	IP        string  `json:"ip"`
	Timestamp int64   `json:"timestamp"`
}

//EventResponse is used as the JSON response to the web request. Using pointer to bool since the field is optional
// and should only be included, when there is a corresponding preceding/subsequent access
type EventResponse struct {
	Current                        Geo       `json:"currentGeo"`
	TravelToCurrentGeoSuspicious   *bool     `json:"travelToCurrentGeoSuspicious,omitempty"`
	TravelFromCurrentGeoSuspicious *bool     `json:"travelFromCurrentGeoSuspicious,omitempty"`
	PrecedingIPAccess              *IPAccess `json:"precedingIpAccess,omitempty"`
	SubsequentIPAccess             *IPAccess `json:"subsequentIpAccess,omitempty"`
}

//Render satisfies the Renderer interface in Chi
func (e *EventResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

//Record is what we are storing in the database
type Record struct {
	ID        int64  `db:"id"`
	UserName  string `db:"username" dynamo:"username" json:"username"`
	Timestamp int64  `db:"timestamp" dynamo:"ts" json:"timestamp"`
	IP        string `db:"ip" json:"ip"`
	Anonymous bool   `db:"anonymous" json:"anonymous"`
	Geo
}

//NewRecord creates a new record, adding the anonymous field
func NewRecord(userName string, timestamp int64, ip string, anonymous bool, lat float64, lon float64, radius uint16) *Record {
	return &Record{
		UserName:  userName,
		Timestamp: timestamp,
		IP:        ip,
		Anonymous: anonymous,
		Geo: Geo{
			Lat:    lat,
			Lon:    lon,
			Radius: radius,
		}}
}
