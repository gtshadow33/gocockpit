package system

import (
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
)

type Service struct {
	Name   string
	Status string
}

type ServiceManager struct {
	systemd *Systemd

	mu       sync.RWMutex
	services map[string]Service
}

func NewServiceManager(systemd *Systemd) *ServiceManager {

	return &ServiceManager{
		systemd:  systemd,
		services: make(map[string]Service),
	}
}

func (m *ServiceManager) loadServices() error {

	var units []struct {
		Name        string
		Description string
		LoadState   string
		ActiveState string
		SubState    string
		Follow      string
		Path        dbus.ObjectPath
		JobID       uint32
		JobType     string
		JobPath     dbus.ObjectPath
	}

	err := m.systemd.obj.Call(
		systemdIface+".Manager.ListUnits",
		0,
	).Store(&units)

	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, unit := range units {

		if !strings.HasSuffix(unit.Name, ".service") {
			continue
		}

		m.services[unit.Name] = Service{
			Name:   unit.Name,
			Status: unit.ActiveState,
		}
	}

	return nil
}

func (m *ServiceManager) GetServices() ([]Service, error) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make([]Service, 0, len(m.services))

	for _, service := range m.services {
		services = append(services, service)
	}

	return services, nil
}

func (m *ServiceManager) StartService(name string) error {

	var jobPath dbus.ObjectPath

	err := m.systemd.obj.Call(
		systemdIface+".Manager.StartUnit",
		0,
		name,
		"replace",
	).Store(&jobPath)

	return err
}

func (m *ServiceManager) StopService(name string) error {

	var jobPath dbus.ObjectPath

	err := m.systemd.obj.Call(
		systemdIface+".Manager.StopUnit",
		0,
		name,
		"replace",
	).Store(&jobPath)

	return err
}

var (
	systemd      *Systemd
	serviceManager *ServiceManager
)

func InitServices() error {

	var err error

	systemd, err = NewSystemd()
	if err != nil {
		return err
	}

	serviceManager = NewServiceManager(systemd)

	return serviceManager.loadServices()
}

func GetServices() ([]Service, error) {

	return serviceManager.GetServices()
}

func StartService(name string) error {

	return serviceManager.StartService(name)
}

func StopService(name string) error {

	return serviceManager.StopService(name)
}