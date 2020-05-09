package geoclue2

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	dbus "github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

const (
	clientPathPrefix = "/org/freedesktop/GeoClue2/Client/"
)

func body(elements ...interface{}) []interface{} {
	ret := make([]interface{}, len(elements))
	for i, e := range elements {
		ret[i] = e
	}
	return ret
}

// func (g *GeoClue2) getClient() error
func TestGetClient(t *testing.T) {
	manager := &MockBusObject{
		DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
			return &dbus.Call{
				Destination: "org.freedesktop.GeoClue2",
				Path:        "/org/freedesktop/GeoClue2/Manager",
				Method:      "GetClient",
				Args:        nil,
				Done:        nil,
				Err:         nil,
				Body:        body("/org/freedesktop/GeoClue2/Client/10"),
			}
		},
	}
	client := &MockBusObject{
		DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
			return &dbus.Call{}
		},
	}
	conn := &MockDbusConn{
		DoObject: func(iface string, path dbus.ObjectPath) dbus.BusObject {
			if path == managerPath {
				return manager
			} else if strings.HasPrefix(string(path), clientPathPrefix) {
				return client
			}
			t.Error(fmt.Sprintf("invalid path %q", path))
			return nil
		},
	}
	gc2 := GeoClue2{
		conn: conn,
	}
	err := gc2.getClient()
	assert.NoError(t, err)
	assert.NotNil(t, gc2.client)
}

func TestGetClientManagerCallErr(t *testing.T) {
	manager := &MockBusObject{
		DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
			return &dbus.Call{
				Err: fmt.Errorf("testing manager.Call() error"),
			}
		},
	}
	client := &MockBusObject{
		DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
			return &dbus.Call{}
		},
	}
	conn := &MockDbusConn{
		DoObject: func(iface string, path dbus.ObjectPath) dbus.BusObject {
			if path == managerPath {
				return manager
			} else if strings.HasPrefix(string(path), clientPathPrefix) {
				return client
			}
			t.Error(fmt.Sprintf("invalid path %q", path))
			return nil
		},
	}
	gc2 := GeoClue2{
		conn: conn,
	}
	err := gc2.getClient()
	assert.Error(t, err)
	assert.Nil(t, gc2.client)
}

func TestGetClientClientCallErr(t *testing.T) {
	manager := &MockBusObject{
		DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
			return &dbus.Call{
				Destination: "org.freedesktop.GeoClue2",
				Path:        "/org/freedesktop/GeoClue2/Manager",
				Method:      "GetClient",
				Args:        nil,
				Done:        nil,
				Err:         nil,
				Body:        body("/org/freedesktop/GeoClue2/Client/10"),
			}
		},
	}
	client := &MockBusObject{
		DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
			return &dbus.Call{
				Err: fmt.Errorf("testing client.Call() error"),
			}
		},
	}
	conn := &MockDbusConn{
		DoObject: func(iface string, path dbus.ObjectPath) dbus.BusObject {
			if path == managerPath {
				return manager
			} else if strings.HasPrefix(string(path), clientPathPrefix) {
				return client
			}
			t.Error(fmt.Sprintf("invalid path %q", path))
			return nil
		},
	}
	gc2 := GeoClue2{
		conn: conn,
	}
	err := gc2.getClient()
	assert.Error(t, err)
	assert.Nil(t, gc2.client)
}

//func (g *GeoClue2) ensureClient() error
func TestEnsureClient(t *testing.T) {
	called := false
	accessed := false
	manager := &MockBusObject{
		DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
			called = true
			return &dbus.Call{
				Destination: "org.freedesktop.GeoClue2",
				Path:        "/org/freedesktop/GeoClue2/Manager",
				Method:      "GetClient",
				Args:        nil,
				Done:        nil,
				Err:         nil,
				Body:        body("/org/freedesktop/GeoClue2/Client/10"),
			}
		},
	}
	client := &MockBusObject{
		DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
			called = true
			return &dbus.Call{}
		},
		DoGetProperty: func(p string) (dbus.Variant, error) {
			accessed = true
			return dbus.MakeVariant(true), nil
		},
	}
	conn := &MockDbusConn{
		DoObject: func(iface string, path dbus.ObjectPath) dbus.BusObject {
			if path == managerPath {
				return manager
			} else if strings.HasPrefix(string(path), clientPathPrefix) {
				return client
			}
			t.Error(fmt.Sprintf("invalid path %q", path))
			return nil
		},
	}
	gc2 := GeoClue2{
		conn: conn,
	}
	err := gc2.ensureClient()
	assert.NoError(t, err)
	assert.NotNil(t, gc2.client)
	assert.True(t, called)
	assert.False(t, accessed)
	called = false
	accessed = false
	err = gc2.ensureClient()
	assert.NoError(t, err)
	assert.NotNil(t, gc2.client)
	assert.False(t, called)
	assert.True(t, accessed)
	called = false
	accessed = false
	client.DoGetProperty = func(p string) (dbus.Variant, error) {
		accessed = true
		return dbus.MakeVariant(false), nil
	}
	err = gc2.ensureClient()
	assert.NoError(t, err)
	assert.NotNil(t, gc2.client)
	assert.True(t, called)
	assert.True(t, accessed)
}

//func getObjInto(intf string, obj dbus.BusObject, into interface{}) error
func TestGetObjInto(t *testing.T) {
	type Embed struct {
		U1 uint64
		U2 uint64
	}
	type Strct struct {
		I     int     `dbus:"I"`
		S     string  `dbus:"S"`
		F     float64 `dbus:"F"`
		E     Embed   `dbus:"E"`
		NoTag string
	}
	strct1 := Strct{
		I: 123456789,
		S: "my-string",
		F: math.Pi,
		E: Embed{1, 2},
	}
	obj := MockBusObject{
		DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
			switch args[1] {
			case "I":
				return &dbus.Call{
					Body: body(dbus.MakeVariant(strct1.I)),
				}
			case "S":
				return &dbus.Call{
					Body: body(dbus.MakeVariant(strct1.S)),
				}
			case "F":
				return &dbus.Call{
					Body: body(dbus.MakeVariant(strct1.F)),
				}
			case "E":
				return &dbus.Call{
					Body: body(dbus.MakeVariant(strct1.E)),
				}
			}
			return nil
		},
	}
	strct2 := Strct{}
	err := getObjInto("", &obj, &strct2)
	assert.NoError(t, err)
	assert.Equal(t, strct1, strct2)
}

func dbusCall(x interface{}) *dbus.Call {
	body := make([]interface{}, 1)
	body[0] = x
	return &dbus.Call{
		Body: body,
	}
}

func mockDbusConn(t *testing.T, manager *MockBusObject, client *MockBusObject, location *MockBusObject) *MockDbusConn {
	if manager == nil {
		manager = &MockBusObject{
			DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
				return &dbus.Call{
					Destination: "org.freedesktop.GeoClue2",
					Path:        "/org/freedesktop/GeoClue2/Manager",
					Method:      "GetClient",
					Body:        body("/org/freedesktop/GeoClue2/Client/10"),
				}
			},
		}
	}
	if client == nil {
		client = &MockBusObject{
			DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
				return &dbus.Call{}
			},
			DoGetProperty: func(p string) (dbus.Variant, error) {
				if p == clientLocation {
					return dbus.MakeVariant(dbus.ObjectPath("location-path")), nil
				}
				return dbus.MakeVariant(true), nil
			},
		}
	}
	if location == nil {
		location = &MockBusObject{
			DoCall: func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
				if len(args) < 2 {
					return &dbus.Call{}
				}
				switch args[1].(string) {
				case "Latitude":
					return dbusCall(1.23)
				case "Longitude":
					return dbusCall(1.23)
				case "Accuracy":
					return dbusCall(1.23)
				case "Altitude":
					return dbusCall(1.23)
				case "Speed":
					return dbusCall(1.23)
				case "Heading":
					return dbusCall(1.23)
				case "Description":
					return dbusCall("")
				case "Timestamp":
					now := time.Now().UnixNano()
					seconds := now / int64(time.Second)
					microseconds := (now - seconds*int64(time.Second)) / int64(time.Microsecond)
					return dbusCall(Timestamp{
						Seconds:      uint64(seconds),
						Microseconds: uint64(microseconds),
					})
				}
				return &dbus.Call{}
			},
			DoGetProperty: func(p string) (dbus.Variant, error) {
				return dbus.MakeVariant(true), nil
			},
		}
	}
	conn := &MockDbusConn{
		DoObject: func(iface string, path dbus.ObjectPath) dbus.BusObject {
			if path == managerPath {
				return manager
			} else if strings.HasPrefix(string(path), clientPathPrefix) {
				return client
			} else if path == "location-path" {
				return location
			}
			t.Error(fmt.Sprintf("invalid path %q", path))
			return nil
		},
		DoSignal: func(ch chan<- *dbus.Signal) {
		},
	}
	return conn
}

func TestStartStop(t *testing.T) {
	gc := GeoClue2{
		conn:        mockDbusConn(t, nil, nil, nil),
		wg:          sync.WaitGroup{},
		quit:        make(chan interface{}),
		dbus:        make(chan *dbus.Signal),
		subscribe:   make(chan chan Location),
		unsubscribe: make(chan chan Location),
	}
	gc.Start()
	gc.Stop()
}

func TestLocationUpdated(t *testing.T) {
	gc := GeoClue2{
		conn:        mockDbusConn(t, nil, nil, nil),
		wg:          sync.WaitGroup{},
		quit:        make(chan interface{}),
		dbus:        make(chan *dbus.Signal),
		subscribe:   make(chan chan Location),
		unsubscribe: make(chan chan Location),
	}
	gc.Start()
	gc.dbus <- &dbus.Signal{
		Name: locationUpdated,
	}
	gc.Stop()
}

func TestGetLocation(t *testing.T) {
	gc := GeoClue2{
		conn:        mockDbusConn(t, nil, nil, nil),
		wg:          sync.WaitGroup{},
		quit:        make(chan interface{}),
		dbus:        make(chan *dbus.Signal),
		subscribe:   make(chan chan Location),
		unsubscribe: make(chan chan Location),
	}
	gc.Start()
	gc.dbus <- &dbus.Signal{
		Name: locationUpdated,
	}
	loc := gc.GetLatestLocation()
	assert.NotNil(t, loc)
	gc.Stop()
}

func TestWaitForLocation(t *testing.T) {
	gc := GeoClue2{
		conn:        mockDbusConn(t, nil, nil, nil),
		wg:          sync.WaitGroup{},
		quit:        make(chan interface{}),
		dbus:        make(chan *dbus.Signal),
		subscribe:   make(chan chan Location),
		unsubscribe: make(chan chan Location),
	}
	gc.Start()
	quit := make(chan interface{})
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		for {
			select {
			case <-quit:
				ticker.Stop()
				return
			case <-ticker.C:
				gc.dbus <- &dbus.Signal{
					Name: locationUpdated,
				}
			}
		}
	}()
	loc, err := gc.WaitForLocation(context.TODO())
	quit <- struct{}{}
	assert.NotNil(t, loc)
	assert.NoError(t, err)
	gc.Stop()
}
