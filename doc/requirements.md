# GoCockpit — Requirements

## Software

* Linux.
* Go 1.27 o superior.
* Git.
* Navegador web moderno.

## Backend

* Go.
* `net/http`.
* `html/template`.
* Go Modules.

## Frontend

* HTML.
* CSS.
* JavaScript.

No se utilizará un framework frontend SPA.

## Sistema

GoCockpit necesita acceso al sistema Linux para obtener información y realizar operaciones.

Debe poder trabajar con:

* `/proc`
* systemd
* Filesystem
* Interfaces de red
* journald
* PAM

## Permisos

Algunas funcionalidades requieren permisos elevados.

GoCockpit deberá utilizar únicamente los permisos necesarios para cada operación.

## Hardware

No se requiere hardware específico.

Los recursos necesarios dependerán del sistema Linux donde se ejecute GoCockpit.

## Desarrollo

Para ejecutar el proyecto:

```bash
go run ./cmd/server
```

Para comprobar que todos los paquetes compilan:

```bash
go build ./...
```

Para crear el ejecutable:

```bash
go build -o gocockpit ./cmd/server
```
