package websocket

import (
	"net/http"

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

	updates := monitor.Connect()
	defer monitor.Disconnect(updates)


	// Esperar actualizaciones del monitor.
	for stats := range updates {

		if err := conn.WriteJSON(stats); err != nil {
			return
		}
	}
}