package monitor

import (
	"sync"
	"time"

	"gocockpit/internal/system"
)

var (
	mu    sync.RWMutex
	stats system.Stats
)

func Start() {

	go func() {

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			newStats := system.GetStats()

			mu.Lock()
			stats = newStats
			mu.Unlock()

			<-ticker.C
		}
	}()

}

func GetStats() system.Stats {

	mu.RLock()
	defer mu.RUnlock()

	return stats
}