package geoclue2

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/godbus/dbus"
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
