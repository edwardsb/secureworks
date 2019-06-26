package httpd

import (
	"context"
	"errors"
	"fmt"
	"github.com/edwardsb/secureworks/geoip"
	"github.com/edwardsb/secureworks/model"
	"github.com/edwardsb/secureworks/store"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/spf13/viper"
	"github.com/umahmood/haversine"
	"log"
	"net"
	"net/http"
	"time"
)

//HTTPServer is the http service
type HTTPServer struct {
	srv     *http.Server
	router  chi.Router
	store   store.Storer
	service geoip.GeoIP
}

//NewHTTPServer is a constructor that will create the HTTPServer with the underlying mux
func NewHTTPServer(storer store.Storer, service geoip.GeoIP) *HTTPServer {

	mux := chi.NewRouter()
	return &HTTPServer{router: mux, store: storer, service: service}

}

//Open will setup routes and start the http server
func (h *HTTPServer) Open() error {
	h.initRouter()
	h.startHTTPServer()
	return nil
}

//Close closes the http server by attempting a graceful shutdown
func (h *HTTPServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := h.srv.Shutdown(ctx)
	if err != nil {
		log.Fatal("failed to shutdown server gracefully")
	}

	log.Println("http server stopped")

	return nil
}

func (h *HTTPServer) initRouter() {

	h.router.Use(HealthCheck("/health"))

	h.router.Route("/v1", func(r chi.Router) {
		r.With(NewEvenRequestMiddleware().Middleware).Post("/", func(w http.ResponseWriter, r *http.Request) {

			request := EvenRequestFromContext(r.Context())
			if request == nil {
				renderError(w, r, errors.New("failed to process request"))
				return
			}

			ip := net.ParseIP(request.IPAddress)

			anonymousIP, err := h.service.AnonymousIP(ip)
			if err != nil {
				log.Printf("failed to determine anonymous ip err: %s\n", err)
				renderError(w, r, err)
				return
			}

			location, err := h.service.Location(ip)
			if err != nil {
				log.Printf("failed to lookup location err: %s\n", err)
				renderError(w, r, err)
				return
			}

			record := model.NewRecord(request.Username,
				request.UnixTimestamp,
				request.IPAddress,
				h.service.IsAnonymous(anonymousIP),
				location.Latitude,
				location.Longitude,
				location.AccuracyRadius)

			_, err = h.store.Put(r.Context(), record)
			if err != nil {
				log.Printf("failed to store event err: %s\n", err)
				renderError(w, r, err)
				return
			}

			response := &model.EventResponse{
				Current: model.Geo{
					Lat:    location.Latitude,
					Lon:    location.Longitude,
					Radius: location.AccuracyRadius,
				},
				TravelToCurrentGeoSuspicious:   nil,
				TravelFromCurrentGeoSuspicious: nil,
				PrecedingIPAccess:              nil,
				SubsequentIPAccess:             nil,
			}

			precedingAccess, err := h.store.PrecedingAccess(r.Context(), request.Username, request.UnixTimestamp)
			if err != nil {
				log.Printf("failed to retrieve current access err: %s\n", err)
				renderError(w, r, err)
				return
			}

			//did we get any preceding login attempts
			if precedingAccess != nil {
				response.TravelToCurrentGeoSuspicious = assignBool(false)
				// since geoip2 returns accuracy radius in km, we have distance in km
				speed, distanceKm := calculateSpeedAndDistance(
					location.Latitude,
					location.Longitude,
					precedingAccess.Lat,
					precedingAccess.Lon,
					precedingAccess.Timestamp,
					request.UnixTimestamp)

				if isSuspicious(speed, distanceKm, location.AccuracyRadius, precedingAccess.Radius) {
					response.TravelToCurrentGeoSuspicious = assignBool(true)
				}
				response.PrecedingIPAccess = &model.IPAccess{
					Geo: model.Geo{
						Lat:    precedingAccess.Lat,
						Lon:    precedingAccess.Lon,
						Radius: precedingAccess.Radius,
					},
					Speed:     speed,
					IP:        precedingAccess.IP,
					Timestamp: precedingAccess.Timestamp,
				}
			}

			subsequentAccess, err := h.store.SubsequentAccess(r.Context(), request.Username, request.UnixTimestamp)
			if err != nil {
				log.Printf("failed to retrieve previous access err: %s\n", err)
				renderError(w, r, err)
				return
			}

			if subsequentAccess != nil {
				response.TravelFromCurrentGeoSuspicious = assignBool(false)
				// since geoip2 returns accuracy radius in km, we have distance in km
				speed, distanceKm := calculateSpeedAndDistance(
					location.Latitude,
					location.Longitude,
					subsequentAccess.Lat,
					subsequentAccess.Lon,
					request.UnixTimestamp,
					subsequentAccess.Timestamp)

				if isSuspicious(speed, distanceKm, location.AccuracyRadius, subsequentAccess.Radius) {
					response.TravelFromCurrentGeoSuspicious = assignBool(true)
				}

				response.SubsequentIPAccess = &model.IPAccess{
					Geo: model.Geo{
						Lat:    subsequentAccess.Lat,
						Lon:    subsequentAccess.Lon,
						Radius: subsequentAccess.Radius,
					},
					Speed:     speed,
					IP:        subsequentAccess.IP,
					Timestamp: subsequentAccess.Timestamp,
				}
			}

			err = render.Render(w, r, response)
			if err != nil {
				renderError(w, r, err)
				return
			}
		})
	})
}

func isSuspicious(speed float64, distance float64, r1, r2 uint16) bool {
	// radius overlap if distance between them is less than the sum of the two radii, although
	// is isn't exactly true for a sphere, because the shortest distance between the center of two circles
	// on a sphere is a straight line, thus under the surface of the sphere. but for short distances on such
	// a big sphere (the earth) I think this will be fine
	// in this case we probably can't accurately enough determine if they are in the overlapping space
	// so it might be better to set not suspicious
	if distance < float64(r1)+float64(r2) {
		return false
	}
	// radii don't overlap, so speed calculation is checked
	return speed > viper.GetFloat64("MAX_SPEED")
}

func calculateSpeedAndDistance(lat1, lon1, lat2, lon2 float64, ts1, ts2 int64) (float64, float64) {
	mi, k := haversine.Distance(haversine.Coord{
		Lat: lat1,
		Lon: lon1,
	}, haversine.Coord{
		Lat: lat2,
		Lon: lon2,
	})

	t1 := time.Unix(ts1, 0)
	t2 := time.Unix(ts2, 0)

	var dur time.Duration
	// we want a positive duration, so always subtract the earlier time, from the later
	if t1.Before(t2) {
		dur = t2.Sub(t1)
	} else {
		dur = t1.Sub(t2)
	}

	//don't divide by zero
	if dur.Hours() > 0 {
		return mi / dur.Hours(), k
	}
	return 0, k
}

//assignBool is a helper to set pointers to bools.
func assignBool(b bool) *bool {
	return &b
}

func renderError(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, http.StatusInternalServerError)
	// definitely not production ready, we could be leaking specifics about our architecture in the form of errors.
	render.JSON(w, r, map[string]interface{}{
		"error": err.Error(),
	})
}

func (h *HTTPServer) startHTTPServer() {

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", 3000),
		Handler: h.router,
	}

	log.Println("starting http server")
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
		}
	}()

	h.srv = srv
}
