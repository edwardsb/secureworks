package httpd

import (
	"context"
	"github.com/edwardsb/secureworks/model"
	"github.com/edwardsb/secureworks/resources"
	"github.com/go-chi/render"
	"github.com/xeipuuv/gojsonschema"
	"log"
	"net/http"
)

//EventRequestValidator is a middleware for validating EventRequests, it requires a schema to be loaded.
type EventRequestValidator struct {
	schema *gojsonschema.Schema
}

//EventKey type to use as context key
type EventKey string

//EventRequestKey is the value of the context key
var EventRequestKey EventKey = "event_request"

//NewEvenRequestMiddleware is a contructor that loads the jsonschema for event requests
func NewEvenRequestMiddleware() *EventRequestValidator {
	l := gojsonschema.NewStringLoader(resources.Get("schemas/eventrequest.json"))
	schema, err := gojsonschema.NewSchema(l)
	if err != nil {
		log.Fatalf("failed to load eventrequest schema, %+v", err)
	}
	return &EventRequestValidator{schema: schema}
}

//Middleware is a middleware function that will bind the request payload to model.EventRequest
//after binding, and validation it will set the event request on the request context.
func (e *EventRequestValidator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var eventRequest = &model.EventRequest{Schema: e.schema}

		err := render.Bind(r, eventRequest)
		if err != nil {
			log.Println(err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]interface{}{
				"errors": err.Error(),
			})
			return
		}

		// dealing with pointers to strings everywhere is kinda messy, so lets just do it once here
		eventRequestValidated := &model.EventRequestValidated{
			Username:      *eventRequest.Username,
			UnixTimestamp: eventRequest.UnixTimestamp,
			EventID:       *eventRequest.EventID,
			IPAddress:     *eventRequest.IPAddress,
		}

		log.Printf("event request: %+v\n", eventRequestValidated)

		//stuff the validated event request in the request context
		ctx := context.WithValue(r.Context(), EventRequestKey, eventRequestValidated)

		//pass along to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//EvenRequestFromContext is simply a convenience method for getting the validated request back out of context
func EvenRequestFromContext(ctx context.Context) *model.EventRequestValidated {
	if m, ok := ctx.Value(EventRequestKey).(*model.EventRequestValidated); ok {
		return m
	}
	return nil
}
