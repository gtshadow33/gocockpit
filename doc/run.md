# Go Build and Project Structure

## Module

GoCockpit utiliza un único módulo definido en `go.mod`:

```go
module gocockpit
```

El módulo contiene todo el proyecto.

```text
gocockpit/
├── go.mod
├── cmd/
├── internal/
└── web/
```

## Packages

Las diferentes carpetas de código pueden contener paquetes.

Por ejemplo:

```text
internal/
├── system/
├── processes/
└── services/
```

Cada paquete puede contener diferentes archivos `.go`.

Los paquetes se pueden importar entre ellos para reutilizar código.

## Main Package

Para crear un programa ejecutable, Go necesita un paquete:

```go
package main
```

y una función:

```go
func main() {
}
```

En GoCockpit se encuentra en:

```text
cmd/server/main.go
```

Este es el punto de entrada del programa.

## `go run`

Para ejecutar GoCockpit directamente:

```bash
go run ./cmd/server
```

Go compila temporalmente el programa y lo ejecuta.

## `go build`

Para crear el ejecutable:

```bash
go build -o gocockpit ./cmd/server
```

Esto genera:

```text
gocockpit
```

El resto de paquetes que utiliza `server` se incluyen automáticamente en el ejecutable.

No estamos compilando solamente la carpeta `server`.

## `go build ./...`

El patrón `./...` significa:

> Este directorio y todos sus subdirectorios.

Por tanto:

```bash
go build ./...
```

compila todos los paquetes del módulo.

Es útil principalmente para comprobar que todo el código del proyecto compila correctamente.

No se utiliza para decidir cuál es el ejecutable principal.

## Diferencia

```bash
go run ./cmd/server
```

Ejecuta GoCockpit.

```bash
go build -o gocockpit ./cmd/server
```

Crea el ejecutable de GoCockpit.

```bash
go build ./...
```

Comprueba que todos los paquetes del proyecto pueden compilarse.

## Resumen

GoCockpit tiene un módulo, varios paquetes y un punto de entrada.

El módulo organiza el proyecto, los paquetes contienen el código y `cmd/server` contiene el `main` que inicia la aplicación.
