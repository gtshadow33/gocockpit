package system

import (
	"github.com/godbus/dbus/v5"
)

const (
	systemdBusName = "org.freedesktop.systemd1"
	systemdPath    = dbus.ObjectPath("/org/freedesktop/systemd1")
	systemdIface   = "org.freedesktop.systemd1"
)

type Systemd struct {
	conn *dbus.Conn
	obj  dbus.BusObject
}

func NewSystemd() (*Systemd, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	obj := conn.Object(
		systemdBusName,
		systemdPath,
	)

	return &Systemd{
		conn: conn,
		obj:  obj,
	}, nil
}

func (s *Systemd) Close() {
	s.conn.Close()
}