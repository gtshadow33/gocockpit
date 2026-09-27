package system

import (
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Service struct {
	Name   string
	Status string
}

var (
	servicesCache []Service
	cacheTime     time.Time

	cacheMu sync.RWMutex

	cacheDuration = 10 * time.Second
)

func GetServices() ([]Service, error) {

	cacheMu.RLock()

	if time.Since(cacheTime) < cacheDuration {

		services := make([]Service, len(servicesCache))
		copy(services, servicesCache)

		cacheMu.RUnlock()

		return services, nil
	}

	cacheMu.RUnlock()

	services, err := loadServices()
	if err != nil {
		return nil, err
	}

	cacheMu.Lock()

	servicesCache = services
	cacheTime = time.Now()

	cacheMu.Unlock()

	return services, nil
}

func loadServices() ([]Service, error) {

	cmd := exec.Command(
		"systemctl",
		"list-units",
		"--type=service",
		"--all",
		"--no-legend",
		"--no-pager",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")

	var services []Service

	for _, line := range lines {

		fields := strings.Fields(line)

		if len(fields) < 4 {
			continue
		}

		name := fields[0]
		status := fields[2]

		if name == "●" {

			if len(fields) < 5 {
				continue
			}

			name = fields[1]
			status = fields[3]
		}

		services = append(services, Service{
			Name:   name,
			Status: status,
		})
	}

	return services, nil
}

func StartService(name string) error {

	cmd := exec.Command(
		"systemctl",
		"start",
		name,
	)

	err := cmd.Run()
	if err != nil {
		return err
	}

	invalidateCache()

	return nil
}

func StopService(name string) error {

	cmd := exec.Command(
		"systemctl",
		"stop",
		name,
	)

	err := cmd.Run()
	if err != nil {
		return err
	}

	invalidateCache()

	return nil
}

func invalidateCache() {

	cacheMu.Lock()
	defer cacheMu.Unlock()

	cacheTime = time.Time{}
	servicesCache = nil
}
