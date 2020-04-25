package geoclue2

import (
	"fmt"
	"strings"
	"testing"

	"github.com/godbus/dbus"
	"github.com/stretchr/testify/assert"
)

const (
	clientPathPrefix = "/org/freedesktop/GeoClue2/Client/"
)

func body(elements []string) []interface{} {
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
				Body:        body([]string{"/org/freedesktop/GeoClue2/Client/10"}),
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
				Body:        body([]string{"/org/freedesktop/GeoClue2/Client/10"}),
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
				Body:        body([]string{"/org/freedesktop/GeoClue2/Client/10"}),
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
