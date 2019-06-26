package geoip

import (
	"github.com/oschwald/geoip2-golang"
	"github.com/pkg/errors"
	"log"
	"net"
)

//GeoIP is an abstraction on GeoIP2-golang. In this case we opt for our own return types
//so later if we decide to switch geoip providers, we just need to supply a new adapter
//from the providers location information, to ours.
type GeoIP interface {
	AnonymousIP(ip net.IP) (*AnonymousIP, error)
	IsAnonymous(ip *AnonymousIP) bool
	Location(ip net.IP) (*Location, error)
}

//Location hold location related data
type Location struct {
	AccuracyRadius uint16
	Latitude       float64
	Longitude      float64
	MetroCode      uint
	TimeZone       string
}

//AnonymousIP holds bools for various types of anonymous ip classifications
type AnonymousIP struct {
	IsAnonymous       bool
	IsAnonymousVPN    bool
	IsHostingProvider bool
	IsPublicProxy     bool
	IsTorExitNode     bool
}

//Service is the wrapper for geoip2 reader, and implements the GeoIP interface
type Service struct {
	reader *geoip2.Reader
	path   string
}

//NewService creates a new GeoIP2 backed Service
func NewService(path string) *Service {
	return &Service{path: path}
}

//AnonymousIP checks GeoIP for Anonymous IPs. Currently this is only supported by the commercial versions of
//GeoIP Anonymous IPs database. For now return false.
func (g *Service) AnonymousIP(ip net.IP) (*AnonymousIP, error) {
	//anonymousIP, err := g.reader.AnonymousIP(ip)
	//if err != nil {
	//	log.Println(err)
	//	return nil, errors.Wrap(err, "failed to lookup anonymous ip")
	//}
	//return &AnonymousIP{
	//	IsAnonymous:       anonymousIP.IsAnonymous,
	//	IsAnonymousVPN:    anonymousIP.IsAnonymousVPN,
	//	IsHostingProvider: anonymousIP.IsHostingProvider,
	//	IsPublicProxy:     anonymousIP.IsPublicProxy,
	//	IsTorExitNode:     anonymousIP.IsTorExitNode,
	//}, nil

	return &AnonymousIP{}, nil
}

//IsAnonymous is just all the Anonymous type OR'd together. Some of these might be less anonymous than others
//but that is business implementation.
func (g *Service) IsAnonymous(ip *AnonymousIP) bool {
	return ip.IsTorExitNode || ip.IsPublicProxy || ip.IsHostingProvider || ip.IsAnonymousVPN || ip.IsAnonymous
}

//Location looks up location for a given IP address, returning the location if available, or and error
func (g *Service) Location(ip net.IP) (*Location, error) {
	city, err := g.reader.City(ip)
	if err != nil {
		return nil, errors.Wrap(err, "failed to lookup city")
	}
	return &Location{
		AccuracyRadius: city.Location.AccuracyRadius,
		Latitude:       city.Location.Latitude,
		Longitude:      city.Location.Longitude,
		MetroCode:      city.Location.MetroCode,
		TimeZone:       city.Location.TimeZone,
	}, nil
}

//Close closes the underlying geoip reader
func (g *Service) Close() error {
	if g.reader != nil {
		err := g.reader.Close()
		if err != nil {
			return err
		}
		log.Println("geoip service closed")
		return nil
	}
	return nil
}

//Open opens the underyling geoip reader
func (g *Service) Open() error {
	if len(g.path) == 0 {
		return errors.New("empty path for geolite db")
	}
	if g.reader == nil {
		reader, err := geoip2.Open(g.path)
		if err != nil {
			return err
		}
		g.reader = reader
	}
	log.Println("opening geoip db")
	return nil
}
