# Decisión: sistema de plantillas HTML

## Contexto

GoCockpit utiliza `html/template` de Go para generar las páginas HTML desde el servidor.

Inicialmente se planteó utilizar un sistema basado en:

* `layout.html`
* `{{ define "content" }}`
* una plantilla global con todas las páginas
* carga automática de todos los archivos HTML

Sin embargo, este enfoque introducía complejidad innecesaria para el tamaño actual del proyecto.

## Decisión

Se decidió utilizar **páginas HTML independientes**, manteniendo únicamente como componentes reutilizables aquellos elementos que realmente se comparten entre páginas.

La estructura actual es:

```text
web/templates/
├── dashboard.html
├── services.html
├── 404.html
├── login.html
├── navbar.html
└── footer.html
```

Las páginas principales son independientes:

```text
dashboard.html
services.html
404.html
login.html
```

Mientras que los elementos comunes se mantienen como componentes:

```text
navbar.html
footer.html
```

Las páginas pueden incluir estos componentes mediante:

```go
{{ template "navbar" . }}
{{ template "footer" . }}
```

No se utiliza un `layout.html` común.

---

## Motivos de la decisión

### 1. Simplicidad

GoCockpit tiene actualmente pocas páginas.

Utilizar un sistema de layouts supondría añadir una capa de abstracción para resolver un problema que todavía no existe.

Con el sistema actual, cada página contiene directamente su estructura HTML:

```html
<!DOCTYPE html>
<html>
<head>
    ...
</head>

<body>

    {{ template "navbar" . }}

    <main>
        ...
    </main>

    {{ template "footer" . }}

</body>
</html>
```

Esto hace que sea sencillo entender qué genera cada controlador.

---

### 2. Evitar conflictos entre plantillas

Un problema del sistema anterior era que varias páginas utilizaban:

```go
{{ define "content" }}
```

y todas las plantillas se cargaban juntas.

Esto podía provocar que una plantilla sobrescribiera el contenido de otra.

Por ejemplo:

```text
dashboard.html
    define "content"

services.html
    define "content"

404.html
    define "content"
```

Al utilizar el mismo nombre dentro de un conjunto de templates, el comportamiento se vuelve más difícil de razonar y mantener.

Con páginas independientes, cada página mantiene su propio contenido y desaparece este problema.

---

### 3. Los componentes comunes siguen siendo reutilizables

Eliminar `layout.html` no significa duplicar todo el HTML.

Los elementos que realmente son compartidos se mantienen como templates independientes:

```text
navbar.html
footer.html
```

Por ejemplo:

```go
{{ template "navbar" . }}
```

y:

```go
{{ template "footer" . }}
```

De esta forma se consigue reutilización sin crear una arquitectura de templates excesivamente compleja.

---

### 4. Cada controlador tiene una responsabilidad clara

El controlador de dashboard renderiza:

```text
dashboard.html
```

El controlador de servicios renderiza:

```text
services.html
```

El controlador de errores renderiza:

```text
404.html
```

Esto permite seguir fácilmente el flujo:

```text
Request
   ↓
Route
   ↓
Controller
   ↓
Template
   ↓
HTML
```

Por ejemplo:

```text
GET /services
      ↓
ServicesController.Index
      ↓
system.GetServices()
      ↓
services.html
```

---

## Alternativas descartadas

### Layout global

Se descartó una estructura como:

```text
layout.html
dashboard.html
services.html
404.html
```

donde todas las páginas utilizan:

```go
{{ define "content" }}
```

porque añade complejidad y puede generar conflictos entre templates cuando se cargan conjuntamente.

---

### Cargar automáticamente todos los HTML

También se descartó recorrer:

```text
web/templates/
```

con `filepath.Walk()` y cargar todos los `.html` mediante:

```go
template.ParseFiles(files...)
```

Este sistema parecía cómodo inicialmente, pero hace que todas las plantillas formen parte del mismo conjunto.

Para GoCockpit esto no aporta una ventaja suficiente y dificulta entender qué templates pertenecen a cada página.

---

### Sistema de componentes más complejo

No se ha introducido ningún motor de templates externo ni un sistema avanzado de componentes.

GoCockpit utiliza directamente:

```go
html/template
```

porque ya proporciona las funcionalidades necesarias.

---

## Ventajas

El sistema actual proporciona:

* poca abstracción;
* pocas dependencias;
* HTML fácil de localizar;
* controladores fáciles de entender;
* componentes reutilizables;
* ausencia de conflictos entre páginas;
* facilidad para modificar una página sin afectar a las demás.

## Inconvenientes

La principal desventaja es que existe cierta repetición de HTML entre páginas, especialmente en:

```html
<!DOCTYPE html>
<html>
<head>
...
</head>
```

Esta duplicación se acepta deliberadamente porque actualmente el número de páginas es reducido.

Si GoCockpit crece significativamente y aparecen muchas páginas con una estructura común, se podrá reconsiderar el uso de un layout.

## Principio aplicado

La decisión sigue el principio:

> **No introducir una abstracción hasta que exista una necesidad real de utilizarla.**

Para el tamaño actual de GoCockpit, un sistema sencillo de páginas independientes y componentes compartidos proporciona un equilibrio adecuado entre reutilización y facilidad de mantenimiento.

Esta decisión puede revisarse si el número de páginas o la complejidad de la interfaz aumenta considerablemente.
