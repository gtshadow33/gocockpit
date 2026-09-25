package monitor

import (
	"sync"
	"time"

	"gocockpit/internal/system"
)

var (
	statsMu sync.RWMutex
	stats   system.Stats

	ctrlMu  sync.Mutex
	clients map[chan system.Stats]struct{}
	stop    chan struct{}
)

func init() {
	clients = make(map[chan system.Stats]struct{})
}

// Connect registra un nuevo WebSocket.
func Connect() chan system.Stats {

	ch := make(chan system.Stats, 1)

	ctrlMu.Lock()
	defer ctrlMu.Unlock()

	clients[ch] = struct{}{}

	if len(clients) == 1 {
		stop = make(chan struct{})
		go run(stop)
	}

	return ch
}

// Disconnect elimina un WebSocket.
func Disconnect(ch chan system.Stats) {

	ctrlMu.Lock()
	defer ctrlMu.Unlock()

	if _, ok := clients[ch]; !ok {
		return
	}

	delete(clients, ch)
	close(ch)

	if len(clients) == 0 {
		close(stop)
	}
}

func run(stop chan struct{}) {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Primera medición inmediatamente.
	update()

	for {
		select {

		case <-stop:
			return

		case <-ticker.C:
			update()
		}
	}
}

func update() {

	newStats := system.GetStats()

	statsMu.Lock()
	stats = newStats
	statsMu.Unlock()

	broadcast(newStats)
}

func broadcast(stats system.Stats) {

	ctrlMu.Lock()
	defer ctrlMu.Unlock()

	for ch := range clients {

		select {
		case ch <- stats:
		default:
			// El cliente todavía no ha consumido
			// la actualización anterior.
		}
	}
}

// GetStats devuelve las últimas estadísticas.
func GetStats() system.Stats {

	statsMu.RLock()
	defer statsMu.RUnlock()

	return stats
}