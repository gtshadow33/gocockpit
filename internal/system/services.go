package system

import (
	"os/exec"
	"strings"
)

type Service struct {
	Name   string
	Status string
}

func GetServices() ([]Service, error) {

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

		// Algunas unidades que no existen
		// aparecen precedidas por "●".
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

	return cmd.Run()
}

func StopService(name string) error {

	cmd := exec.Command(
		"systemctl",
		"stop",
		name,
	)

	return cmd.Run()
}

