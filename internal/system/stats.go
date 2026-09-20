package system

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Stats struct {
	CPU int
	RAM int
}

type cpuStats struct {
	total uint64
	idle  uint64
}

func GetStats() Stats {
	return Stats{
		CPU: getCPUUsage(),
		RAM: getRAMUsage(),
	}
}

func readCPUStats() cpuStats {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return cpuStats{}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	if !scanner.Scan() {
		return cpuStats{}
	}

	fields := strings.Fields(scanner.Text())

	if len(fields) < 5 {
		return cpuStats{}
	}

	var values []uint64

	for _, field := range fields[1:] {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return cpuStats{}
		}

		values = append(values, value)
	}

	var total uint64

	for _, value := range values {
		total += value
	}

	idle := values[3] + values[4]

	return cpuStats{
		total: total,
		idle:  idle,
	}
}

func getCPUUsage() int {
	first := readCPUStats()

	time.Sleep(500 * time.Millisecond)

	second := readCPUStats()

	totalDelta := second.total - first.total
	idleDelta := second.idle - first.idle

	if totalDelta == 0 {
		return 0
	}

	usage := (totalDelta - idleDelta) * 100 / totalDelta

	return int(usage)
}

func getRAMUsage() int {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer file.Close()

	var total uint64
	var available uint64

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			total, _ = strconv.ParseUint(fields[1], 10, 64)

		case "MemAvailable:":
			available, _ = strconv.ParseUint(fields[1], 10, 64)
		}
	}

	if err := scanner.Err(); err != nil {
		return 0
	}

	if total == 0 {
		return 0
	}

	used := total - available

	return int((used * 100) / total)
}