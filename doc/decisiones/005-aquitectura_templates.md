# 005 — Organización y carga de templates

**Estado:** Aceptada
**Fecha:** 25/09/2026

---

## 1. Contexto

GoCockpit utiliza las plantillas HTML del paquete estándar:

```go
html/template
```

Inicialmente las plantillas se cargaban mediante `template.ParseGlob()`.

Esta solución era suficiente mientras todos los archivos HTML estaban directamente dentro de:

```text
web/templates/
```

Sin embargo, se quiere permitir una organización mediante subdirectorios para separar las plantillas por funcionalidad.

Por ejemplo:

```text
web/
└── templates/
    ├── auth/
    ├── dashboard/
    └── errors/
```

---

# 2. Solución anterior

Inicialmente `main.go` utilizaba:

```go
templates := template.Must(
	template.ParseGlob("web/templates/*.html"),
)
```

La estructura esperada era:

```text
web/
└── templates/
    ├── login.html
    ├── dashboard.html
    └── 404.html
```

El patrón:

```text
web/templates/*.html
```

busca archivos `.html` directamente dentro de `templates`.

Por ejemplo:

```text
web/templates/login.html       ← encontrado
web/templates/dashboard.html   ← encontrado
```

Pero si se crea una subcarpeta:

```text
web/
└── templates/
    ├── login.html
    └── auth/
        └── register.html
```

el archivo:

```text
auth/register.html
```

no es encontrado por:

```go
template.ParseGlob("web/templates/*.html")
```

porque el patrón no realiza un recorrido recursivo.

---

# 3. Problema

La solución anterior obligaba a mantener todas las plantillas en una única carpeta.

Con el crecimiento de GoCockpit esto podría producir una carpeta difícil de mantener:

```text
web/templates/
├── login.html
├── register.html
├── dashboard.html
├── settings.html
├── profile.html
├── 404.html
├── 500.html
├── users.html
├── user.html
└── ...
```

Se quiere poder organizar las vistas:

```text
web/templates/
├── auth/
│   ├── login.html
│   └── register.html
│
├── dashboard/
│   ├── index.html
│   └── settings.html
│
└── errors/
    ├── 404.html
    └── 500.html
```

---

# 4. Nueva solución

Se decide utilizar:

```go
filepath.Walk()
```

para recorrer recursivamente:

```text
web/templates/
```

y localizar todos los archivos `.html`.

El código actual es:

```go
var files []string

err := filepath.Walk("web/templates", func(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !info.IsDir() && filepath.Ext(path) == ".html" {
		files = append(files, path)
	}

	return nil
})

if err != nil {
	log.Fatal(err)
}

templates := template.Must(
	template.ParseFiles(files...),
)
```

---

# 5. Diferencia entre ambas soluciones

### Antes

```text
ParseGlob()
     │
     ▼
web/templates/*.html
     │
     ├── login.html
     ├── dashboard.html
     └── 404.html
```

Solo busca directamente en `templates`.

---

### Ahora

```text
filepath.Walk()
        │
        ▼
web/templates/
        │
        ├── login.html
        │
        ├── auth/
        │   └── register.html
        │
        ├── dashboard/
        │   └── index.html
        │
        └── errors/
            └── 404.html
        │
        ▼
    ParseFiles()
        │
        ▼
 *template.Template
```

Ahora se recorren las subcarpetas automáticamente.

---

# 6. Código completo anterior

El `main.go` anterior era:

```go
package main

import (
	"html/template"
	"log"
	"net/http"

	"gocockpit/internal/system"
	"gocockpit/routes"
)

func main() {

	templates := template.Must(
		template.ParseGlob("web/templates/*.html"),
	)

	mux := http.NewServeMux()

	routes.Register(mux, templates)
	system.Start()

	log.Println("GoCockpit escuchando en http://localhost:8080")

	log.Fatal(
		http.ListenAndServe(":8080", mux),
	)
}
```

---

# 7. Código completo nuevo

El `main.go` actual es:

```go
package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gocockpit/internal/system"
	"gocockpit/routes"
)

func main() {

	var files []string

	err := filepath.Walk("web/templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".html" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	templates := template.Must(
		template.ParseFiles(files...),
	)

	mux := http.NewServeMux()

	routes.Register(mux, templates)
	system.Start()

	log.Println("GoCockpit escuchando en http://localhost:8080")

	log.Fatal(
		http.ListenAndServe(":8080", mux),
	)
}
```

---

# 8. Funcionamiento del nuevo código

Primero se crea un slice:

```go
var files []string
```

Este slice almacenará las rutas de las plantillas encontradas.

Después:

```go
filepath.Walk("web/templates", ...)
```

comienza a recorrer:

```text
web/templates/
```

y todas sus subcarpetas.

Para cada archivo se ejecuta:

```go
if !info.IsDir() && filepath.Ext(path) == ".html"
```

Esto significa:

1. Comprobar que no sea una carpeta.
2. Comprobar que tenga extensión `.html`.

Si cumple ambas condiciones:

```go
files = append(files, path)
```

se añade su ruta al slice.

Finalmente:

```go
template.ParseFiles(files...)
```

carga todas las plantillas encontradas.

---

# 9. Ventaja arquitectónica

La principal ventaja es que los controladores no necesitan conocer cómo están organizadas físicamente las plantillas.

`main.go` se encarga de descubrirlas:

```text
main.go
   │
   └── descubre templates
             │
             ▼
      *template.Template
             │
             ▼
      routes.Register()
             │
       ┌─────┼─────┐
       ▼     ▼     ▼
      Auth Dashboard Errors
```

Los controladores simplemente reciben:

```go
Templates *template.Template
```

y utilizan las plantillas.

---

# 10. Estructura resultante

La estructura de vistas puede crecer de esta forma:

```text
web/
└── templates/
    ├── auth/
    │   ├── login.html
    │   └── register.html
    │
    ├── dashboard/
    │   ├── index.html
    │   └── settings.html
    │
    ├── errors/
    │   ├── 404.html
    │   └── 500.html
    │
    └── users/
        ├── index.html
        └── show.html
```

No es necesario modificar el mecanismo de carga cada vez que se crea una nueva carpeta.

---

# 11. Decisión final

Se abandona:

```go
template.ParseGlob("web/templates/*.html")
```

como mecanismo de descubrimiento de plantillas.

Se adopta:

```go
filepath.Walk()
```

para descubrir recursivamente los archivos `.html` y:

```go
template.ParseFiles()
```

para cargarlos.

La decisión permite mantener las plantillas organizadas por funcionalidad y mantiene la responsabilidad de carga de recursos en el punto de entrada de la aplicación.
