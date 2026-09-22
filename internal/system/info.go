package system

import (
	"net"
	"os"
	"runtime"
	"strings"
)

type SystemInfo struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	CPUs     int    `json:"cpus"`
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
}

var info SystemInfo

func Start() {
	hostname, _ := os.Hostname()

	ip, mac := getNetworkInfo()

	info = SystemInfo{
		Hostname: hostname,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPUs:     runtime.NumCPU(),
		IP:       ip,
		MAC:       mac,
	}
}

func GetInfo() SystemInfo {
	return info
}

func getNetworkInfo() (string, string) {

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", ""
	}

	for _, iface := range interfaces {

		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, address := range addresses {

			ip := strings.Split(address.String(), "/")[0]

			parsedIP := net.ParseIP(ip)

			if parsedIP != nil && parsedIP.To4() != nil {
				return ip, iface.HardwareAddr.String()
			}
		}
	}

	return "", ""
}