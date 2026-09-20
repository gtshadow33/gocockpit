# Initial Project Structure

## Structure

GoCockpit utiliza una estructura organizada por responsabilidades.

* `cmd/server/`: punto de entrada del servidor.
* `internal/`: lógica interna de la aplicación.
* `web/templates/`: plantillas HTML.
* `web/static/`: archivos CSS y JavaScript.
* `docs/`: documentación del proyecto.

El servidor se inicia desde `cmd/server/main.go` y utiliza `net/http`.
