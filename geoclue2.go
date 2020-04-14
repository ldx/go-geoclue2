package geoclue2

import (
	"context"
	"reflect"
	"sync"

	"github.com/godbus/dbus"
	"k8s.io/klog"
)

const (
	defaultDesktopID  = "go-geoclue2"
	getProperties     = "org.freedesktop.DBus.Properties.Get"
	setProperties     = "org.freedesktop.DBus.Properties.Set"
	geoClue2Interface = "org.freedesktop.GeoClue2"
	clientInterface   = "org.freedesktop.GeoClue2.Client"
	clientActive      = "org.freedesktop.GeoClue2.Client.Active"
	clientLocation    = "org.freedesktop.GeoClue2.Client.Location"
	clientStart       = "org.freedesktop.GeoClue2.Client.Start"
	locationUpdated   = "org.freedesktop.GeoClue2.Client.LocationUpdated"
	locationInterface = "org.freedesktop.GeoClue2.Location"
	getClient         = "org.freedesktop.GeoClue2.Manager.GetClient"
	managerPath       = "/org/freedesktop/GeoClue2/Manager"
)

// This is the location as returned by GeoClue2.
type Location struct {
	// The latitude of the location, in degrees.
	Latitude float64 `dbus:"Latitude"`
	// The longitude of the location, in degrees.
	Longitude float64 `dbus:"Longitude"`
	// The accuracy of the location fix, in meters.
	Accuracy float64 `dbus:"Accuracy"`
	// The altitude of the location fix, in meters. When unknown, its set to
	// minimum double value, -1.7976931348623157e+308.
	Altitude float64 `dbus:"Altitude"`
	// The speed in meters per second. When unknown, it's set to -1.0.
	Speed float64 `dbus:"Speed"`
	// The heading direction in degrees with respect to North direction, in
	// clockwise order. That means North becomes 0 degree, East: 90 degrees,
	// South: 180 degrees, West: 270 degrees and so on. When unknown, it's set
	// to -1.0.
	Heading float64 `dbus:"Heading"`
	// A human-readable description of the location, if available.
	// WARNING: Applications should not rely on this property since not all
	// sources provide a description. If you really need a description (or more
	// details) about current location, use a reverse-geocoding API, e.g
	// geocode-glib.
	Description string `dbus:"Description"`
	// The timestamp when the location was determined, in seconds and
	// microseconds since the Epoch. This is the time of measurement if the
	// backend provided that information, otherwise the time when GeoClue
	// received the new location. Note that GeoClue can't guarantee that the
	// timestamp will always monotonically increase, as a backend may not
	// respect that. Also note that a timestamp can be very old, e.g. because
	// of a cached location.
	Timestamp []uint64 `dbus:"Timestamp"`
}

type GeoClue2 struct {
	conn           *dbus.Conn
	desktopID      string
	wg             sync.WaitGroup
	quit           chan interface{}
	dbusCh         chan *dbus.Signal
	subscribe      chan chan Location
	unsubscribe    chan chan Location
	client         dbus.BusObject
	latestLocation *Location
}

func NewGeoClue2(conn *dbus.Conn, desktopID string) *GeoClue2 {
	if desktopID == "" {
		desktopID = defaultDesktopID
	}
	dbusCh := make(chan *dbus.Signal)
	conn.Signal(dbusCh)
	return &GeoClue2{
		conn:        conn,
		desktopID:   desktopID,
		wg:          sync.WaitGroup{},
		quit:        make(chan interface{}),
		dbusCh:      dbusCh,
		subscribe:   make(chan chan Location),
		unsubscribe: make(chan chan Location),
	}
}

func (g *GeoClue2) Start() {
	go g.controlLoop()
}

func (g *GeoClue2) Stop() {
	g.quit <- struct{}{}
	g.wg.Wait()
}

func (g *GeoClue2) ensureClient() error {
	if g.client == nil {
		return g.getClient()
	}
	active := false
	value, err := g.client.GetProperty(clientActive)
	if err != nil {
		return g.getClient()
	}
	active = value.Value().(bool)
	if !active {
		return g.getClient()
	}
	return nil
}

func (g *GeoClue2) getClient() error {
	var path string
	manager := g.conn.Object(geoClue2Interface, managerPath)
	call := manager.Call(getClient, 0)
	err := call.Store(&path)
	if err != nil {
		klog.Warningf("getting client: %v", err)
		return err
	}
	clientPath := dbus.ObjectPath(path)
	client := g.conn.Object(geoClue2Interface, clientPath)
	id := dbus.MakeVariant(g.desktopID)
	err = client.Call(setProperties, 0, clientInterface, "DesktopId", id).Err
	if err != nil {
		klog.Warningf("setting DesktopId: %v", err)
		return err
	}
	err = client.Call(clientStart, 0).Err
	if err != nil {
		klog.Warningf("starting client: %v", err)
		return err
	}
	g.client = client
	return nil
}

func (g *GeoClue2) getObjInto(intf string, obj dbus.BusObject, into interface{}) error {
	s := reflect.ValueOf(into).Elem()
	t := s.Type()
	for i := 0; i < t.NumField(); i++ {
		structF := t.Field(i)
		fName, ok := structF.Tag.Lookup("dbus")
		if !ok {
			continue
		}
		f := s.Field(i)
		if f.Kind() == reflect.Slice {
			var tmp []interface{}
			err := obj.Call(getProperties, 0, intf, fName).Store(&tmp)
			if err != nil {
				return err
			}
			f.Set(reflect.MakeSlice(f.Type(), len(tmp), cap(tmp)))
			for j, tmpval := range tmp {
				idx := f.Index(j)
				idx.Set(reflect.ValueOf(tmpval).Convert(idx.Type()))
			}
		} else {
			fAddr := f.Addr().Interface()
			err := obj.Call(getProperties, 0, intf, fName).Store(fAddr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *GeoClue2) processLocationUpdate() *Location {
	val, err := g.client.GetProperty(clientLocation)
	if err != nil {
		klog.Warningf("getting location path from update: %v", err)
		return nil
	}
	path := val.Value().(dbus.ObjectPath)
	obj := g.conn.Object("org.freedesktop.GeoClue2", path)
	location := Location{}
	err = g.getObjInto(locationInterface, obj, &location)
	if err != nil {
		klog.Warningf("getting location object from update: %v", err)
		return nil
	}
	return &location
}

func (g *GeoClue2) GetLatestLocation() *Location {
	return g.latestLocation
}

func (g *GeoClue2) WaitForLocation(ctx context.Context) (*Location, error) {
	ch := make(chan Location)
	g.subscribe <- ch
	select {
	case loc := <-ch:
		g.unsubscribe <- ch
		return &loc, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (g *GeoClue2) broadcastUpdate(subscribers map[chan<- Location]interface{}, loc Location) {
	for ch, _ := range subscribers {
		select {
		case ch <- loc:
		default:
		}
	}
}

func (g *GeoClue2) controlLoop() {
	g.wg.Add(1)
	defer g.wg.Done()
	klog.V(2).Infof("starting up")
	subscribers := make(map[chan<- Location]interface{})
	for {
		g.ensureClient()
		select {
		case subscribe := <-g.subscribe:
			klog.V(5).Infof("new subscriber")
			subscribers[subscribe] = ""
		case unsubscribe := <-g.unsubscribe:
			klog.V(5).Infof("subscriber gone")
			delete(subscribers, unsubscribe)
		case sig := <-g.dbusCh:
			klog.V(5).Infof("dbus signal %s", sig.Name)
			if sig.Name == locationUpdated {
				loc := g.processLocationUpdate()
				if loc != nil {
					g.latestLocation = loc
					g.broadcastUpdate(subscribers, *loc)
				}
			}
		case <-g.quit:
			klog.V(2).Infof("shutting down")
			return
		}
	}
}
