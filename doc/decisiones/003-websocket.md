# WebSocket para estadísticas en tiempo real

## Objetivo

Mostrar las estadísticas de CPU y RAM del servidor en el dashboard sin tener que recargar la página.

Para ello se utiliza un WebSocket entre Go y el navegador.

---

## Arquitectura

El sistema está dividido en tres partes:

```text
/proc
  ↓
system
  ↓
monitor
  ↓
WebSocket
  ↓
JavaScript
  ↓
Dashboard
```

### `system`

Se encarga de leer las estadísticas reales de Linux.

Lee:

* `/proc/stat` para CPU.
* `/proc/meminfo` para RAM.

Devuelve:

```go
type Stats struct {
	CPU int `json:"cpu"`
	RAM int `json:"ram"`
}
```

Los `json` tags hacen que Go envíe:

```json
{
    "cpu": 35,
    "ram": 62
}
```

en lugar de:

```json
{
    "CPU": 35,
    "RAM": 62
}
```

---

## Monitor

El paquete `monitor` mantiene las estadísticas actuales en memoria.

Su objetivo es que las estadísticas se calculen **una sola vez** y puedan ser utilizadas por múltiples conexiones WebSocket.

```go
var (
	mu    sync.RWMutex
	stats system.Stats
)
```

El mutex protege el acceso a `stats`.

* `Lock()` se utiliza cuando se actualiza.
* `RLock()` se utiliza cuando se lee.

---

## Goroutine del monitor

El monitor se inicia desde `main.go`:

```go
monitor.Start()
```

`Start()` crea una goroutine que actualiza las estadísticas periódicamente:

```go
func Start() {

	go func() {

		ticker := time.NewTicker(1 * time.Second)
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
```

De esta forma existe una única tarea encargada de calcular las estadísticas.

Esto evita que cada usuario tenga que ejecutar individualmente `system.GetStats()`.

---

## Obtener las estadísticas

Los WebSockets utilizan:

```go
func GetStats() system.Stats {

	mu.RLock()
	defer mu.RUnlock()

	return stats
}
```

Esto permite que varias conexiones puedan leer las estadísticas compartidas.

---

## WebSocket

El endpoint utilizado es:

```text
/ws/stats
```

La ruta se registra en `routes.go`:

```go
mux.HandleFunc("/ws/stats", websocket.Stats)
```

Cuando el navegador se conecta a `/ws/stats`, se ejecuta:

```go
websocket.Stats
```

---

## Conexión WebSocket

El servidor actualiza la conexión HTTP a WebSocket:

```go
conn, err := upgrader.Upgrade(w, r, nil)
if err != nil {
	return
}
```

Después se mantiene la conexión abierta:

```go
defer conn.Close()
```

---

## Envío de estadísticas

El WebSocket obtiene las estadísticas del monitor:

```go
stats := monitor.GetStats()
```

y las envía como JSON:

```go
err := conn.WriteJSON(stats)
if err != nil {
	return
}
```

Se utiliza un ticker para enviar los datos periódicamente:

```go
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()
```

El código completo queda:

```go
func Stats(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer conn.Close()

	ticker := time.NewTicker(1 * time.Second)
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
```

---

## JavaScript

El navegador crea la conexión:

```javascript
const protocol =
    window.location.protocol === "https:"
        ? "wss:"
        : "ws:";

const socket = new WebSocket(
    `${protocol}//${window.location.host}/ws/stats`
);
```

Se utilizan:

* `ws://` cuando la página utiliza HTTP.
* `wss://` cuando la página utiliza HTTPS.

---

## Recibir los datos

Cuando llega un mensaje:

```javascript
socket.onmessage = function (event) {

    const stats = JSON.parse(event.data);

};
```

Por ejemplo, si Go envía:

```json
{
    "cpu": 43,
    "ram": 67
}
```

JavaScript obtiene:

```javascript
stats.cpu // 43
stats.ram // 67
```

---

## Actualización del dashboard

Los elementos HTML tienen IDs:

```html
<span id="cpu-header"></span>

<div id="cpu-value"></div>

<div id="cpu-bar" class="bar-fill"></div>
```

JavaScript modifica su contenido:

```javascript
document.getElementById("cpu-header").textContent =
    `${stats.cpu}%`;

document.getElementById("cpu-value").textContent =
    `${stats.cpu}%`;
```

La barra se modifica cambiando su ancho:

```javascript
document.getElementById("cpu-bar").style.width =
    `${stats.cpu}%`;
```

Por ejemplo:

```javascript
stats.cpu = 40;
```

produce:

```css
width: 40%;
```

La barra pasa a ocupar el 40% de su contenedor.

---

## Flujo completo

Cuando arranca GoCockpit:

```text
main.go
   ↓
monitor.Start()
   ↓
goroutine del monitor
   ↓
system.GetStats()
   ↓
actualiza stats
```

Cuando entra un usuario:

```text
Navegador
   ↓
/dashboard
   ↓
HTML
```

Después JavaScript abre:

```text
/ws/stats
   ↓
websocket.Stats
   ↓
monitor.GetStats()
   ↓
WriteJSON()
   ↓
JavaScript
```

Finalmente JavaScript actualiza:

```text
CPU %
RAM %
barra CPU
barra RAM
```

---

## Ventaja de este diseño

Las estadísticas no se calculan individualmente para cada usuario.

Con varios usuarios:

```text
                 ┌── WebSocket usuario 1
                 │
monitor ─────────┼── WebSocket usuario 2
                 │
                 ├── WebSocket usuario 3
                 │
                 └── WebSocket usuario N
```

El monitor calcula las estadísticas y todos los clientes reciben el último estado disponible.

Esto separa las responsabilidades:

* `system` → obtiene datos de Linux.
* `monitor` → mantiene el estado actual.
* `websocket` → distribuye los datos.
* JavaScript → actualiza la interfaz.

---

## Dependencia

Se utiliza la librería:

```bash
go get github.com/gorilla/websocket
```

El import utilizado es:

```go
import "github.com/gorilla/websocket"
```

---

## Archivos implicados

```text
gocockpit/
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── monitor/
│   │   └── monitor.go
│   │
│   ├── system/
│   │   └── stats.go
│   │
│   └── websocket/
│       └── stats.go
│
├── routes/
│   └── routes.go
│
└── web/
    └── templates/
        └── dashboard.html
```

Cada archivo tiene una responsabilidad concreta y el WebSocket no se encarga directamente de calcular CPU o RAM.

## Decisión

Se utiliza WebSocket para las estadísticas que necesitan actualización en tiempo real, manteniendo el dashboard como HTML renderizado por el servidor.

No se utiliza JavaScript para solicitar continuamente `/dashboard` ni se recarga la página para actualizar los valores.
