# Architecture

GoCockpit utiliza una arquitectura web sencilla basada en Go.

* **Frontend:** HTML, CSS y JavaScript.
* **Backend:** Go con `net/http`.
* **Handlers:** gestionan las peticiones.
* **Services:** contienen la lógica.
* **Linux:** proporciona la información y operaciones del sistema.
* **WebSocket:** utilizado para funcionalidades en tiempo real.

La aplicación se desarrollará de forma modular, añadiendo funcionalidades progresivamente.
