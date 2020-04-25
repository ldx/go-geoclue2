package geoclue2

import "github.com/godbus/dbus"

type MockBusObject struct {
	DoCall        func(method string, flags dbus.Flags, args ...interface{}) *dbus.Call
	DoGo          func(method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call
	DoGetProperty func(p string) (dbus.Variant, error)
	DoDestination func() string
	DoPath        func() dbus.ObjectPath
}

func (o *MockBusObject) Call(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	return o.DoCall(method, flags, args...)
}

func (o *MockBusObject) Go(method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	return o.DoGo(method, flags, ch, args...)
}

func (o *MockBusObject) GetProperty(p string) (dbus.Variant, error) {
	return o.DoGetProperty(p)
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
