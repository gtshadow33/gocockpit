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
	clients int
	stop    chan struct{}
)

// Connect debe llamarse cuando un cliente WS se conecta.
// Arranca el polling solo si es el primer cliente.
func Connect() {
	ctrlMu.Lock()
	defer ctrlMu.Unlock()

	clients++

	if clients == 1 {
		stop = make(chan struct{})
		go run(stop)
	}
}

// Disconnect debe llamarse cuando un cliente WS se desconecta.
// Para el polling si era el último cliente.
func Disconnect() {
	ctrlMu.Lock()
	defer ctrlMu.Unlock()

	if clients == 0 {
		return
	}

	clients--

	if clients == 0 {
		close(stop)
	}
}

func run(stop chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

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
}

// GetStats devuelve el último valor cacheado (no calcula nada).
func GetStats() system.Stats {
	statsMu.RLock()
	defer statsMu.RUnlock()

	return stats
}