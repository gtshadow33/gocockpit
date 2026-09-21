package websocket

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"gocockpit/internal/monitor"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Stats(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer conn.Close()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {

		stats := monitor.GetStats()

		err := conn.WriteJSON(stats)
		if err != nil {
			return
		}

		<-ticker.C
	}
}