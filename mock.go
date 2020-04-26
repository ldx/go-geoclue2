package geoclue2

import (
	"context"

	dbus "github.com/godbus/dbus/v5"
)

type MockBusObject struct {
	DoCall              func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call
	DoCallWithContext   func(ctx context.Context, method string, flags dbus.Flags, args ...interface{}) *dbus.Call
	DoGo                func(method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call
	DoGoWithContext     func(ctx context.Context, method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call
	DoAddMatchSignal    func(iface, member string, options ...dbus.MatchOption) *dbus.Call
	DoRemoveMatchSignal func(iface, member string, options ...dbus.MatchOption) *dbus.Call
	DoGetProperty       func(p string) (dbus.Variant, error)
	DoSetProperty       func(p string, v interface{}) error
	DoDestination       func() string
	DoPath              func() dbus.ObjectPath
}

func (o *MockBusObject) Call(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	return o.DoCall(method, flags, args...)
}

func (o *MockBusObject) CallWithContext(ctx context.Context, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	return o.DoCallWithContext(ctx, method, flags, args...)
}

func (o *MockBusObject) Go(method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	return o.DoGo(method, flags, ch, args...)
}

func (o *MockBusObject) GoWithContext(ctx context.Context, method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	return o.DoGoWithContext(ctx, method, flags, ch, args...)
}

func (o *MockBusObject) AddMatchSignal(iface, member string, options ...dbus.MatchOption) *dbus.Call {
	return o.DoAddMatchSignal(iface, member, options...)
}

func (o *MockBusObject) RemoveMatchSignal(iface, member string, options ...dbus.MatchOption) *dbus.Call {
	return o.DoRemoveMatchSignal(iface, member, options...)
}

func (o *MockBusObject) GetProperty(p string) (dbus.Variant, error) {
	return o.DoGetProperty(p)
}

func (o *MockBusObject) SetProperty(p string, v interface{}) error {
	return o.DoSetProperty(p, v)
}

func (o *MockBusObject) Destination() string {
	return o.DoDestination()
}

func (o *MockBusObject) Path() dbus.ObjectPath {
	return o.DoPath()
}

type MockDbusConn struct {
	DoSignal func(ch chan<- *dbus.Signal)
	DoObject func(iface string, path dbus.ObjectPath) dbus.BusObject
}

func (d *MockDbusConn) Signal(ch chan<- *dbus.Signal) {
	d.DoSignal(ch)
}

func (d *MockDbusConn) Object(iface string, path dbus.ObjectPath) dbus.BusObject {
	return d.DoObject(iface, path)
}
