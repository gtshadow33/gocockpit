# Decisión 006 — Monitor centralizado y broadcast de estadísticas

**Archivo:** `doc/desicions/006-monitor-websocket.md`

**Estado:** Aceptada

El monitor centraliza la obtención periódica de estadísticas del sistema y las distribuye mediante broadcast a todos los WebSockets conectados.

No se crea un `ticker` por WebSocket. Existe un único `ticker` de 10 segundos en `monitor`, que ejecuta `system.GetStats()` una vez y distribuye el resultado a todos los clientes.

```text
                    monitor
                       │
                  ticker 10s
                       │
                       ▼
               system.GetStats()
                       │
                       ▼
                  broadcast()
                ┌──────┼──────┐
                ▼      ▼      ▼
               WS1    WS2    WS3
```

El monitor comienza a funcionar cuando se conecta el primer cliente y se detiene cuando se desconecta el último.

Cada WebSocket dispone de su propio canal para recibir las actualizaciones, evitando que los clientes compitan por los mensajes de un canal compartido.

```text
monitor
   │
   ├── canal ──► WS1
   ├── canal ──► WS2
   └── canal ──► WS3
```

De esta forma:

* `system` obtiene las estadísticas.
* `monitor` controla el polling, almacena el estado y realiza el broadcast.
* `websocket` gestiona únicamente las conexiones y transmite los datos al navegador.
* Todos los clientes reciben la misma muestra de estadísticas.
* `system.GetStats()` se ejecuta una sola vez cada 10 segundos, independientemente del número de clientes.
* El polling se detiene cuando no existen clientes conectados.

La decisión establece que:

> **El monitor es la única fuente de actualización de estadísticas y los WebSockets son consumidores de esas actualizaciones.**

Esto evita cálculos duplicados, mantiene todos los clientes sincronizados y separa claramente las responsabilidades entre `system`, `monitor` y `websocket`.
