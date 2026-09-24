# Monitorización bajo demanda y caché

## Descripción

Se ha modificado el sistema de monitorización de GoCockpit para evitar que el servidor esté obteniendo estadísticas del sistema continuamente cuando no existen clientes conectados mediante WebSocket.

Anteriormente, el monitor ejecutaba una goroutine permanente que obtenía las estadísticas cada cierto tiempo, independientemente de si algún cliente las necesitaba.

La nueva implementación utiliza un **contador de clientes WebSocket** para controlar el ciclo de vida del proceso de monitorización.

El funcionamiento pasa a ser:

```text
                    Cliente WebSocket
                           │
                           ▼
                    monitor.Connect()
                           │
                    ¿Es el primero?
                       │         │
                      NO        SÍ
                       │         │
                       │         ▼
                       │    Inicia polling
                       │         │
                       └─────────┤
                                 ▼
                         Actualiza caché
                                 │
                                 ▼
                         Cliente consulta
                         monitor.GetStats()
```

Cuando se desconecta el último cliente:

```text
Último cliente
      │
      ▼
monitor.Disconnect()
      │
      ▼
clients == 0
      │
      ▼
Detener polling
```

---

# 1. Separación entre información estática y monitorización

La obtención de información estática del sistema continúa siendo independiente del monitor.

`system.Start()` se ejecuta una vez durante el arranque del servidor:

```text
Inicio GoCockpit
       │
       ▼
system.Start()
       │
       ▼
Obtiene información estática
       │
       ▼
Guarda SystemInfo
```

Esta información puede incluir datos como:

* Hostname
* Sistema operativo
* Arquitectura
* Número de CPUs
* IP
* MAC

Estos datos no necesitan actualizarse continuamente.

Por tanto:

```text
system.Start()
    │
    └── Información estática
         └── Se obtiene al arrancar
```

Mientras que `monitor` se ocupa de las estadísticas dinámicas:

```text
monitor
   │
   └── CPU
   └── RAM
       └── Se actualizan periódicamente
```

---

# 2. Caché de estadísticas

El monitor mantiene una única copia de las últimas estadísticas obtenidas.

```go
var (
    statsMu sync.RWMutex
    stats   system.Stats
)
```

La variable `stats` representa la caché.

La función `update()` obtiene nuevas estadísticas y sustituye el valor anterior:

```go
func update() {
    newStats := system.GetStats()

    statsMu.Lock()
    stats = newStats
    statsMu.Unlock()
}
```

La función `GetStats()` **no calcula estadísticas nuevas**.

Simplemente devuelve el último valor almacenado:

```go
func GetStats() system.Stats {
    statsMu.RLock()
    defer statsMu.RUnlock()

    return stats
}
```

Esto permite que varios clientes consulten simultáneamente las mismas estadísticas.

---

# 3. Contador de clientes

El monitor mantiene un contador de conexiones WebSocket activas:

```go
var (
    ctrlMu  sync.Mutex
    clients int
    stop    chan struct{}
)
```

`clients` indica cuántos clientes WebSocket están utilizando actualmente el monitor.

Por ejemplo:

```text
0 clientes → monitor detenido

1 cliente  → monitor activo

2 clientes → monitor activo

3 clientes → monitor activo

0 clientes → monitor detenido
```

El contador está protegido mediante `ctrlMu` para evitar condiciones de carrera cuando varios clientes se conectan o desconectan simultáneamente.

---

# 4. Inicio del polling

Cuando un cliente WebSocket se conecta, se ejecuta:

```go
monitor.Connect()
```

La función incrementa el contador:

```go
clients++
```

Si ese cliente es el primero:

```go
if clients == 1 {
    stop = make(chan struct{})
    go run(stop)
}
```

Por tanto, **solo el primer cliente inicia la goroutine de monitorización**.

El flujo es:

```text
WebSocket
    │
    ▼
monitor.Connect()
    │
    ▼
clients++
    │
    ▼
¿clients == 1?
    │
   SÍ
    │
    ▼
Crear canal stop
    │
    ▼
Crear goroutine
    │
    ▼
run()
```

---

# 5. Polling

La función `run()` es la encargada de realizar el polling:

```go
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
```

El funcionamiento es:

1. Se crea un `Ticker` de 10 segundos.
2. Se realiza una actualización inmediatamente.
3. Se espera al siguiente tick.
4. Se vuelven a obtener las estadísticas.
5. El proceso continúa mientras existan clientes.

Es importante que se haga `update()` inmediatamente al arrancar.

De esta forma, el primer cliente no tiene que esperar 10 segundos para recibir datos.

```text
Connect()
   │
   ▼
run()
   │
   ├── update() inmediatamente
   │
   ▼
esperar 10 segundos
   │
   ▼
update()
   │
   ▼
esperar 10 segundos
   │
   ▼
update()
```

---

# 6. Detención del polling

Cuando un cliente WebSocket se desconecta se ejecuta:

```go
monitor.Disconnect()
```

La función reduce el contador:

```go
clients--
```

Si todavía existen clientes:

```text
clients > 0
```

el monitor continúa funcionando.

Si se ha desconectado el último:

```text
clients == 0
```

se cierra el canal:

```go
close(stop)
```

Esto provoca que la goroutine salga del `select`:

```go
case <-stop:
    return
```

El flujo completo es:

```text
Cliente desconectado
        │
        ▼
monitor.Disconnect()
        │
        ▼
clients--
        │
        ▼
¿clients == 0?
      │       │
     NO      SÍ
      │       │
      │       ▼
      │   close(stop)
      │       │
      │       ▼
      │   run() termina
      │
      ▼
Monitor continúa
```

---

# 7. Funcionamiento con varios clientes

Todos los clientes utilizan la misma caché.

Por ejemplo, si hay tres clientes:

```text
             ┌──────────────┐
             │   MONITOR    │
             │              │
             │ 1 goroutine  │
             │              │
             │    CACHE     │
             └──────┬───────┘
                    │
          ┌─────────┼─────────┐
          │         │         │
          ▼         ▼         ▼
       Cliente 1 Cliente 2 Cliente 3
```

No se crean tres procesos de monitorización.

Solo existe una goroutine realizando:

```text
system.GetStats()
```

Los clientes simplemente leen:

```text
monitor.GetStats()
```

Esto evita realizar el mismo trabajo varias veces.

---

# 8. WebSocket

El controlador WebSocket ahora controla el ciclo de vida del monitor:

```go
monitor.Connect()
defer monitor.Disconnect()
```

El uso de `defer` garantiza que `Disconnect()` se ejecute cuando termina la función, independientemente de si la conexión termina por un error o por una desconexión normal.

El flujo es:

```text
Conexión WebSocket
        │
        ▼
Upgrade()
        │
        ▼
monitor.Connect()
        │
        ▼
Enviar estadísticas
        │
        ▼
Cliente sigue conectado
        │
        ├── cada 10 segundos
        │       │
        │       ▼
        │   GetStats()
        │       │
        │       ▼
        │   WriteJSON()
        │
        ▼
Conexión termina
        │
        ▼
monitor.Disconnect()
```

El WebSocket **no obtiene directamente las estadísticas del sistema**.

Las obtiene de la caché:

```text
WebSocket
    │
    ▼
monitor.GetStats()
    │
    ▼
Caché
```

---

# 9. Nuevo flujo completo

La arquitectura de monitorización queda de esta forma:

```text
                    ┌─────────────────┐
                    │    NAVEGADOR    │
                    └────────┬────────┘
                             │
                       WebSocket
                             │
                             ▼
                    ┌─────────────────┐
                    │    WEBSOCKET    │
                    └────────┬────────┘
                             │
                       Connect()
                             │
                             ▼
                    ┌─────────────────┐
                    │     MONITOR     │
                    │                 │
                    │ contador clientes
                    │ caché           │
                    └────────┬────────┘
                             │
                       1ª conexión
                             │
                             ▼
                    ┌─────────────────┐
                    │    GOROUTINE    │
                    │    POLLING      │
                    └────────┬────────┘
                             │
                       cada 10 segundos
                             │
                             ▼
                    ┌─────────────────┐
                    │     SYSTEM      │
                    │                 │
                    │ /proc/stat      │
                    │ /proc/meminfo   │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │      CACHE      │
                    │    CPU / RAM    │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │    WEBSOCKET    │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │    NAVEGADOR    │
                    └─────────────────┘
```

---

# 10. Cambios realizados en `main.go`

Se elimina el arranque permanente del monitor:

```go
monitor.Start()
```

El monitor ya no necesita iniciarse al arrancar el servidor.

Ahora su ciclo de vida depende de las conexiones WebSocket.

El arranque queda:

```text
main()
 │
 ├── Cargar templates
 │
 ├── Crear ServeMux
 │
 ├── Registrar routes
 │
 ├── system.Start()
 │
 └── ListenAndServe()
```

`system.Start()` permanece porque la información estática sí debe inicializarse al arrancar.

---

# 11. Motivo del cambio

El objetivo principal es evitar trabajo innecesario.

### Antes

```text
Servidor iniciado
      │
      ▼
Monitor siempre activo
      │
      ▼
Cada 10 segundos
      │
      ▼
Obtener CPU/RAM
      │
      ▼
Aunque no haya clientes
```

### Ahora

```text
Servidor iniciado
      │
      ▼
No hay clientes
      │
      ▼
Monitor detenido


Cliente conecta
      │
      ▼
Monitor arranca
      │
      ▼
Polling cada 10 segundos


Último cliente desconecta
      │
      ▼
Monitor se detiene
```

De esta manera, el sistema de monitorización funciona **bajo demanda**.

---

# 12. Resumen de responsabilidades

| Componente             | Responsabilidad                                              |
| ---------------------- | ------------------------------------------------------------ |
| `system.Start()`       | Inicializar información estática del sistema                 |
| `system.GetInfo()`     | Obtener la información estática almacenada                   |
| `system.GetStats()`    | Obtener estadísticas actuales del sistema                    |
| `monitor.Connect()`    | Registrar un cliente y arrancar el polling si es el primero  |
| `monitor.Disconnect()` | Desregistrar un cliente y detener el polling si es el último |
| `monitor.run()`        | Ejecutar el polling periódico                                |
| `monitor.update()`     | Actualizar la caché                                          |
| `monitor.GetStats()`   | Leer las estadísticas cacheadas                              |
| `websocket.Stats()`    | Gestionar la conexión WebSocket y enviar estadísticas        |
| `main()`               | Inicializar la aplicación y el servidor HTTP                 |

## Resultado

La modificación introduce una separación clara entre:

```text
INFORMACIÓN ESTÁTICA
        │
        └── system.Start()
              └── una vez al arrancar


ESTADÍSTICAS DINÁMICAS
        │
        └── monitor
              ├── polling bajo demanda
              ├── una goroutine compartida
              └── caché compartida


CLIENTES
        │
        └── WebSocket
              └── consumen la caché
```

El resultado es un sistema de monitorización compartido, cacheado y activado únicamente cuando existen clientes WebSocket interesados en las estadísticas.
