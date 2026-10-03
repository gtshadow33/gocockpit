package websocket

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"

	"gocockpit/internal/middleware"
	"gocockpit/internal/terminal"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

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
		// host. Evita que otra web use la cookie del usuario para abrir
		// una shell (Cross-Site WebSocket Hijacking).
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
	// Si no está, la ruta no pasó por Auth y se rechaza antes del upgrade.
	username, ok := r.Context().Value(
		middleware.UsernameKey,
	).(string)

	if !ok || username == "" {
		http.Error(w, "No autenticado", http.StatusUnauthorized)
		return
	}

	log.Println("Terminal: usuario:", username)

	// Upgrade HTTP -> WebSocket. Si falla, el Upgrader ya respondió
	// al cliente con el error HTTP correspondiente.
	conn, err := terminalUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Terminal: error WebSocket:", err)
		return
	}

	// Se cierra la conexión al salir del handler, pase lo que pase.
	defer conn.Close()

	log.Println("Terminal: WebSocket conectado")

	// Lanza /bin/bash con los permisos del usuario autenticado dentro de
	// una PTY. Requiere que GoCockpit corra como root (o con CAP_SETUID
	// y CAP_SETGID) para cambiar de uid/gid.
	term, err := terminal.Start(username)
	if err != nil {
		log.Println("Terminal: error iniciando PTY:", err)

		// Avisa al usuario en pantalla antes de cerrar.
		conn.WriteMessage(
			websocket.TextMessage,
			[]byte("Error iniciando terminal: "+err.Error()+"\r\n"),
		)

		return
	}

	// Al terminar la sesión se cierra la PTY y se mata la shell.
	defer term.Close()

	log.Println("Terminal: PTY iniciada")

	// --- PTY -> WebSocket ---------------------------------------------
	// Goroutine que lee la salida de la shell y la envía al navegador.
	go func() {

		// Si la shell termina (por ejemplo con "exit"), la lectura falla
		// y se cierra el WebSocket. Eso desbloquea el ReadMessage del
		// bucle principal y termina el handler.
		defer conn.Close()

		buffer := make([]byte, 4096)

		for {

			n, err := term.PTY.Read(buffer)

			// Se procesan primero los bytes leídos: Read puede devolver
			// datos y un error en la misma llamada.
			if n > 0 {

				// Se envía como binario porque un chunk puede cortar un
				// carácter UTF-8 a la mitad, y el navegador cerraría el
				// socket si lo recibiera como texto inválido.
				if werr := conn.WriteMessage(
					websocket.BinaryMessage,
					buffer[:n],
				); werr != nil {
					log.Println("Terminal: error escribiendo WebSocket:", werr)
					return
				}
			}

			if err != nil {

				// io.EOF es el cierre normal; cualquier otro error se registra.
				if err != io.EOF {
					log.Println("Terminal: error leyendo PTY:", err)
				}

				return
			}
		}
	}()

	// --- WebSocket -> PTY ---------------------------------------------
	// Bucle principal: lee lo que envía el navegador.
	// Protocolo:
	//   - Mensaje de texto:   teclas pulsadas, se escriben tal cual en la PTY.
	//   - Mensaje binario:    JSON {"cols":N,"rows":M} para redimensionar.
	for {

		msgType, data, err := conn.ReadMessage()
		if err != nil {
			// El cliente cerró la pestaña, o la goroutine cerró la conexión.
			log.Println("Terminal: WebSocket cerrado:", err)
			return
		}

		if msgType == websocket.BinaryMessage {

			var size resizeMessage

			// Se ignoran mensajes mal formados o con tamaño 0.
			if json.Unmarshal(data, &size) == nil &&
				size.Cols > 0 && size.Rows > 0 {

				// Informa a la PTY del nuevo tamaño. Así vim, htop o bash
				// ajustan su salida al tamaño real de la ventana.
				pty.Setsize(term.PTY, &pty.Winsize{
					Cols: size.Cols,
					Rows: size.Rows,
				})
			}

			continue
		}

		// Teclas del usuario -> entrada estándar de la shell.
		if _, err := term.PTY.Write(data); err != nil {
			log.Println("Terminal: error escribiendo PTY:", err)
			return
		}
	}
}