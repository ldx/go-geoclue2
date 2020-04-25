package geoclue2

import (
	"github.com/godbus/dbus"
)

type DbusConn interface {
	Signal(ch chan<- *dbus.Signal)
	Object(iface string, path dbus.ObjectPath) dbus.BusObject
}

type RealDbusConn struct {
	conn *dbus.Conn
}

func (d *RealDbusConn) Signal(ch chan<- *dbus.Signal) {
	d.conn.Signal(ch)
}

func (d *RealDbusConn) Object(iface string, path dbus.ObjectPath) dbus.BusObject {
	return d.conn.Object(iface, path)
}
