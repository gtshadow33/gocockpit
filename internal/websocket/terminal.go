package websocket

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"gocockpit/internal/middleware"
	"gocockpit/internal/session"
	"gocockpit/internal/terminal"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

const (
	// Máximo de terminales abiertas a la vez por usuario.
	maxTerminalsPerUser = 5

	// Cada pingPeriod se envía un ping y se revisa sesión e inactividad.
	pingPeriod = 30 * time.Second

	// Si no llega ningún pong/mensaje en pongWait, la conexión se da por muerta.
	pongWait = 70 * time.Second

	// Sin teclas del usuario durante este tiempo, se cierra la terminal.
	idleTimeout = 30 * time.Minute

	// Tamaño máximo de un mensaje del navegador.
	maxMessageSize = 64 * 1024
)

var (
	activeMu sync.Mutex
	active   = make(map[string]int)
)

// acquire reserva un hueco de terminal para el usuario.
func acquire(username string) bool {

	activeMu.Lock()
	defer activeMu.Unlock()

	if active[username] >= maxTerminalsPerUser {
		return false
	}

	active[username]++

	return true
}

// release libera el hueco reservado con acquire.
func release(username string) {

	activeMu.Lock()
	defer activeMu.Unlock()

	active[username]--

	if active[username] <= 0 {
		delete(active, username)
	}
}

// Upgrader convierte la petición HTTP en una conexión WebSocket.
// CheckOrigin decide si se acepta el origen de la petición.
var terminalUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {

		origin := r.Header.Get("Origin")

		// Clientes que no son navegador (curl, scripts) no envían Origin.
		// Siguen protegidos por la cookie de sesión del middleware Auth.
		if origin == "" {
			return true
		}

		u, err := url.Parse(origin)
		if err != nil {
			return false
		}

		// Solo se acepta si la página que abre el socket es de este mismo
		// host. Evita Cross-Site WebSocket Hijacking.
		return u.Host == r.Host
	},
}

// Mensaje JSON que envía el cliente cuando cambia el tamaño del terminal.
type resizeMessage struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// Terminal gestiona una sesión de terminal por WebSocket.
// Por cada conexión se crea una shell con su propia PTY.
func Terminal(w http.ResponseWriter, r *http.Request) {

	// El middleware Auth guarda el usuario como string en el contexto.
	username, ok := r.Context().Value(
		middleware.UsernameKey,
	).(string)

	if !ok || username == "" {
		http.Error(w, "No autenticado", http.StatusUnauthorized)
		return
	}

	// ID de sesión: se revisa periódicamente para cortar la terminal
	// si el usuario hace logout o la sesión caduca.
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "No autenticado", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value

	// Límite de terminales simultáneas por usuario.
	if !acquire(username) {
		http.Error(
			w,
			"Demasiadas terminales abiertas",
			http.StatusTooManyRequests,
		)
		return
	}
	defer release(username)

	log.Println("Terminal: usuario:", username)

	conn, err := terminalUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Terminal: error WebSocket:", err)
		return
	}

	defer conn.Close()

	log.Println("Terminal: WebSocket conectado")

	// Lanza bash con uid, gid y grupos del usuario autenticado.
	term, err := terminal.Start(username)
	if err != nil {
		log.Println("Terminal: error iniciando PTY:", err)

		conn.WriteMessage(
			websocket.TextMessage,
			[]byte("Error iniciando terminal: "+err.Error()+"\r\n"),
		)

		return
	}

	defer term.Close()

	log.Println("Terminal: PTY iniciada")

	// --- Keepalive ----------------------------------------------------
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))

	// El navegador responde a los ping con pong automáticamente.
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Última vez que el usuario escribió algo (nanosegundos Unix).
	var lastActivity atomic.Int64
	lastActivity.Store(time.Now().UnixNano())

	// done se cierra al salir del handler y para el supervisor.
	done := make(chan struct{})
	defer close(done)

	// --- Supervisor ---------------------------------------------------
	// Cada pingPeriod: comprueba sesión, inactividad y envía ping.
	// WriteControl es seguro de llamar a la vez que WriteMessage.
	go func() {

		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {

			case <-done:
				return

			case <-ticker.C:

				if _, valid := session.Get(sessionID); !valid {
					log.Println("Terminal: sesión caducada o cerrada:", username)
					closeWithReason(conn, "sesión caducada")
					return
				}

				idle := time.Since(time.Unix(0, lastActivity.Load()))

				if idle > idleTimeout {
					log.Println("Terminal: cerrada por inactividad:", username)
					closeWithReason(conn, "inactividad")
					return
				}

				if err := conn.WriteControl(
					websocket.PingMessage,
					nil,
					time.Now().Add(10*time.Second),
				); err != nil {
					conn.Close()
					return
				}
			}
		}
	}()

	// --- PTY -> WebSocket ---------------------------------------------
	go func() {

		// Si la shell termina, se cierra el WebSocket y eso desbloquea
		// el ReadMessage del bucle principal.
		defer conn.Close()

		buffer := make([]byte, 4096)

		for {

			n, err := term.PTY.Read(buffer)

			if n > 0 {

				// Binario: un chunk puede cortar un carácter UTF-8.
				if werr := conn.WriteMessage(
					websocket.BinaryMessage,
					buffer[:n],
				); werr != nil {
					log.Println("Terminal: error escribiendo WebSocket:", werr)
					return
				}
			}

			if err != nil {

				if err != io.EOF {
					log.Println("Terminal: error leyendo PTY:", err)
				}

				return
			}
		}
	}()

	// --- WebSocket -> PTY ---------------------------------------------
	// Protocolo:
	//   - Mensaje de texto:   teclas pulsadas, se escriben en la PTY.
	//   - Mensaje binario:    JSON {"cols":N,"rows":M} para redimensionar.
	for {

		msgType, data, err := conn.ReadMessage()
		if err != nil {
			log.Println("Terminal: WebSocket cerrado:", err)
			return
		}

		// Cualquier mensaje prueba que el cliente está vivo.
		conn.SetReadDeadline(time.Now().Add(pongWait))

		if msgType == websocket.BinaryMessage {

			var size resizeMessage

			if json.Unmarshal(data, &size) == nil &&
				size.Cols > 0 && size.Rows > 0 {

				pty.Setsize(term.PTY, &pty.Winsize{
					Cols: size.Cols,
					Rows: size.Rows,
				})
			}

			continue
		}

		// Solo las teclas cuentan como actividad (no los resize).
		lastActivity.Store(time.Now().UnixNano())

		if _, err := term.PTY.Write(data); err != nil {
			log.Println("Terminal: error escribiendo PTY:", err)
			return
		}
	}
}

// closeWithReason envía un frame de cierre con motivo y cierra la conexión.
func closeWithReason(conn *websocket.Conn, reason string) {

	conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason),
		time.Now().Add(5*time.Second),
	)

	conn.Close()
}