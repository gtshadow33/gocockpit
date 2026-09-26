# Sistema de configuración

## Objetivo

GoCockpit necesita una configuración externa para evitar tener valores importantes escritos directamente en el código.

Actualmente se configurarán:

* Host del servidor HTTP.
* Puerto del servidor HTTP.
* Grupo Linux que tiene permisos de administrador.

## Formato

Se utilizará **TOML** como formato de configuración.

Ejemplo:

```toml
host = "0.0.0.0"
port = 8080
admin_group = "wheel"
```

TOML permite mantener una configuración sencilla y legible sin tener que implementar un parser propio.

## Ubicación

Durante el desarrollo, el archivo se encuentra en la raíz del proyecto:

```text
gocockpit/
├── gocockpit.toml
├── cmd/
├── internal/
├── routes/
└── web/
```

En una instalación definitiva se podrá trasladar a una ubicación del sistema, por ejemplo:

```text
/etc/gocockpit/gocockpit.toml
```

## Estructura en Go

La configuración pertenece al paquete:

```text
internal/config/
└── config.go
```

La estructura utilizada es:

```go
type Config struct {
    Host       string `toml:"host"`
    Port       int    `toml:"port"`
    AdminGroup string `toml:"admin_group"`
}
```

La configuración se mantiene en:

```go
var App Config
```

y se carga al iniciar GoCockpit mediante:

```go
config.Load("gocockpit.toml")
```

## Flujo

El proceso de arranque es:

```text
                    gocockpit.toml
                           │
                           ▼
                    config.Load()
                           │
                           ▼
                       config.App
                       /        \
                      /          \
                     ▼            ▼
                HTTP server    Middleware
                                │
                                ▼
                           AdminGroup
```

El `main.go` es responsable de cargar la configuración.

Los diferentes paquetes pueden acceder a los valores que necesitan mediante:

```go
config.App
```

## Motivo de esta decisión

No se pasará la configuración manualmente a través de todas las capas de la aplicación.

Por ejemplo, las rutas no necesitan recibir:

```go
routes.Register(mux, templates, cfg)
```

Se mantiene:

```go
routes.Register(mux, templates)
```

Esto evita que `routes` tenga que conocer información que no necesita.

El middleware de administración puede acceder directamente a:

```go
config.App.AdminGroup
```

De esta forma, la responsabilidad queda separada:

* `main` → carga la configuración.
* `config` → representa y gestiona la configuración.
* `routes` → registra las rutas.
* `middleware` → utiliza los valores de configuración que necesita.
* `system` → monitoriza el sistema.

## Compatibilidad entre distribuciones

El grupo administrativo puede variar según la distribución Linux.

Por ejemplo:

### Arch Linux

```toml
admin_group = "wheel"
```

### Fedora

```toml
admin_group = "wheel"
```

### Debian / Ubuntu

```toml
admin_group = "sudo"
```

No es necesario modificar el código de Go para cambiar de distribución. Solamente se modifica:

```toml
admin_group
```

## Ejemplo completo

### gocockpit.toml

```toml
host = "0.0.0.0"
port = 8080
admin_group = "wheel"
```

### Carga

```go
if err := config.Load("gocockpit.toml"); err != nil {
    log.Fatal("Error cargando configuración: ", err)
}
```

### Uso

```go
address := config.App.Host + ":" + strconv.Itoa(config.App.Port)
```

Y en el middleware:

```go
if group == config.App.AdminGroup {
    next.ServeHTTP(w, r)
    return
}
```

## Decisión

Se utilizará **TOML + `internal/config`** para gestionar la configuración de GoCockpit.

La configuración se cargará una única vez durante el arranque de la aplicación y estará disponible para los componentes que necesiten utilizarla.
