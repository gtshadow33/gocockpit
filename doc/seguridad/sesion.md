# GoCockpit — Arquitectura y sistema de autenticación

## 1. Objetivo actual

GoCockpit es una aplicación web escrita en Go para administrar un sistema Linux desde una interfaz web.

El objetivo actual es construir la aplicación **sin framework web**, utilizando principalmente la librería estándar de Go para entender cómo funciona internamente una aplicación HTTP.

Actualmente se ha construido la base de:

* Servidor HTTP.
* Routing.
* Controllers.
* Middleware.
* Autenticación mediante PAM.
* Sesiones.
* Cookies de sesión.
* Contextos de Go.
* Separación de responsabilidades.

---

# 2. Arquitectura actual

La estructura del proyecto es:

```text
gocockpit/
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── auth/
│   │   └── auth.go
│   │
│   ├── controllers/
│   │   ├── auth.go
│   │   └── dashboard.go
│   │
│   ├── middleware/
│   │   └── auth.go
│   │
│   └── session/
│       └── session.go
│
├── routes/
│   └── routes.go
│
├── web/
│   ├── static/
│   │   ├── css/
│   │   └── js/
│   │
│   └── templates/
│
└── doc/
    └── ...
```

Cada parte tiene una responsabilidad concreta.

---

# 3. Responsabilidades

## `cmd/server`

Es el punto de entrada de la aplicación.

Se encarga de arrancar:

* Templates.
* Router.
* Servidor HTTP.

No contiene la lógica de autenticación ni de negocio.

---

## `routes`

Relaciona las URLs con sus handlers.

Ejemplo:

```text
/login      → AuthController.Login
/dashboard  → Auth Middleware → DashboardController.Index
```

El routing decide **qué código se ejecuta para cada URL**.

---

## `controllers`

Los controllers reciben las peticiones HTTP y ejecutan la lógica correspondiente.

Por ejemplo:

```go
func (d *DashboardController) Index(
    w http.ResponseWriter,
    r *http.Request,
)
```

El controller no debería encargarse de comprobar directamente las cookies si esa responsabilidad ya pertenece al middleware.

---

# 4. Autenticación mediante PAM

El paquete:

```text
internal/auth/
```

se encarga exclusivamente de autenticar las credenciales.

La función principal es:

```go
func Authenticate(username, password string) error
```

Utiliza:

```go
github.com/msteinert/pam/v2
```

El flujo es:

```text
username + password
        │
        ▼
auth.Authenticate()
        │
        ▼
       PAM
        │
   ┌────┴────┐
   │         │
 válido    inválido
   │         │
   ▼         ▼
 nil       error
```

El paquete `auth` no sabe nada de:

* Cookies.
* HTTP.
* Sessions.
* Controllers.
* Context.

Su única responsabilidad es **comprobar las credenciales**.

---

# 5. Sesiones

El paquete:

```text
internal/session/
```

se encarga de mantener las sesiones.

Actualmente se utiliza un mapa protegido mediante `sync.RWMutex`:

```go
var (
    sessions = make(map[string]string)
    mu       sync.RWMutex
)
```

La relación es:

```text
sessionID → username
```

Por ejemplo:

```text
a3f91c7e... → pepe
```

## Crear sesión

```go
func Create(username string) (string, error)
```

Genera 32 bytes aleatorios:

```go
bytes := make([]byte, 32)
```

Utiliza:

```go
crypto/rand
```

y después los convierte a hexadecimal:

```go
sessionID := hex.EncodeToString(bytes)
```

Como cada byte hexadecimal ocupa dos caracteres:

```text
32 bytes × 2 = 64 caracteres
```

Por tanto, un ID puede tener este aspecto:

```text
a3f91c7e8b2d4a...e91f0c2a
```

El ID es aleatorio y no contiene directamente el nombre del usuario.

Después se almacena:

```go
sessions[sessionID] = username
```

---

# 6. Cookies

Después de autenticar al usuario y crear la sesión, el servidor envía una cookie:

```go
http.SetCookie(w, &http.Cookie{
    Name:     "session_id",
    Value:    sessionID,
    Path:     "/",
    HttpOnly: true,
    SameSite: http.SameSiteStrictMode,
})
```

El navegador almacena:

```text
session_id = a3f91c7e...
```

En las siguientes peticiones al servidor, el navegador envía automáticamente:

```http
Cookie: session_id=a3f91c7e...
```

El servidor puede utilizar ese ID para buscar la sesión:

```text
Cookie
  │
  ▼
session_id
  │
  ▼
session.Get()
  │
  ▼
username
```

---

# 7. `HttpOnly`

La cookie utiliza:

```go
HttpOnly: true
```

Esto evita que JavaScript pueda acceder directamente a ella mediante:

```javascript
document.cookie
```

Pero el navegador sigue enviándola automáticamente en las peticiones HTTP correspondientes.

Por tanto:

```text
JavaScript ❌ → leer cookie
Navegador   ✅ → enviar cookie al servidor
```

---

# 8. `SameSite`

Se utiliza:

```go
SameSite: http.SameSiteStrictMode
```

Esto añade restricciones para que la cookie no se envíe en determinados contextos de navegación cross-site.

---

# 9. Middleware de autenticación

El middleware está en:

```text
internal/middleware/auth.go
```

Su responsabilidad es comprobar si una petición tiene una sesión válida.

El flujo es:

```text
Request
   │
   ▼
¿Existe session_id?
   │
   ├── NO → /login
   │
   ▼
session.Get(sessionID)
   │
   ├── NO EXISTE → /login
   │
   ▼
Sesión válida
   │
   ▼
Controller
```

El middleware no autentica la contraseña mediante PAM.

Eso pertenece a `auth`.

El middleware trabaja con la sesión que ya fue creada después del login.

---

# 10. ¿Por qué el middleware obtiene el username?

Cuando hacemos:

```go
username, exists := session.Get(cookie.Value)
```

obtenemos el usuario asociado a la sesión.

Podríamos hacer que el controller volviera a realizar exactamente la misma operación, pero eso provocaría duplicación:

```text
Middleware
    ↓
session.Get()

Controller
    ↓
session.Get()
```

Por eso el middleware puede dejar disponible el usuario para el resto de la petición.

Para ello utilizamos `context.Context`.

---

# 11. Contexto de Go

Un contexto puede almacenar valores asociados a una petición.

Primero definimos una clave:

```go
type contextKey string

const UsernameKey contextKey = "username"
```

`UsernameKey` **no es el usuario**.

Es la clave que utilizamos para localizar el usuario.

Conceptualmente:

```text
UsernameKey
     │
     ▼
   "pepe"
```

Cuando hacemos:

```go
ctx := context.WithValue(
    r.Context(),
    UsernameKey,
    username,
)
```

estamos asociando:

```text
clave             valor

UsernameKey  →  "pepe"
```

Después el contexto se coloca en la petición:

```go
next.ServeHTTP(
    w,
    r.WithContext(ctx),
)
```

El siguiente handler puede recuperarlo:

```go
username := r.Context().Value(UsernameKey)
```

---

# 12. El contexto puede contener varios datos

Un contexto puede tener varias claves:

```go
type contextKey string

const (
    UsernameKey contextKey = "username"
    RoleKey     contextKey = "role"
    SessionKey  contextKey = "session"
)
```

Y podemos guardar:

```go
ctx := context.WithValue(
    r.Context(),
    UsernameKey,
    "pepe",
)

ctx = context.WithValue(
    ctx,
    RoleKey,
    "admin",
)

ctx = context.WithValue(
    ctx,
    SessionKey,
    "abc123",
)
```

Conceptualmente:

```text
Context
│
├── UsernameKey → "pepe"
├── RoleKey     → "admin"
└── SessionKey  → "abc123"
```

El contexto permite que diferentes partes de la misma petición accedan a información que necesitan.

---

# 13. Responsabilidad de cada componente

La separación actual es:

```text
AUTH
 │
 └── ¿Las credenciales son correctas?

SESSION
 │
 └── ¿Qué usuario corresponde a este session_id?

MIDDLEWARE
 │
 └── ¿La petición tiene una sesión válida?
 │
 └── Deja disponible la identidad de la petición.

CONTROLLER
 │
 └── ¿Qué hacemos con esa petición autenticada?

ROUTES
 │
 └── ¿Qué handler corresponde a esta URL?

MAIN
 │
 └── ¿Cómo arrancamos toda la aplicación?
```

Esta separación es importante porque evita que una sola pieza del programa tenga demasiadas responsabilidades.

---

# 14. Flujo completo del login

Cuando el usuario entra en:

```text
GET /login
```

el controller muestra el formulario.

Después:

```text
POST /login
```

El controller recibe:

```text
username
password
```

y ejecuta:

```go
auth.Authenticate(username, password)
```

Si PAM acepta las credenciales:

```text
PAM
 ↓
OK
 ↓
session.Create(username)
 ↓
sessionID
 ↓
Set-Cookie
 ↓
redirect /dashboard
```

---

# 15. Flujo completo de `/dashboard`

El navegador solicita:

```text
GET /dashboard
```

y envía:

```http
Cookie: session_id=abc123...
```

El middleware recibe la petición:

```text
Cookie
  ↓
session_id
  ↓
session.Get()
  ↓
"pepe"
  ↓
context
  ↓
DashboardController
```

El controller puede obtener:

```go
username := r.Context().Value(UsernameKey)
```

y utilizarlo:

```go
fmt.Fprintf(w, "Hola %s", username)
```

---

# 16. Lo que hemos conseguido

La aplicación ya tiene los fundamentos de una aplicación web real:

```text
                    GoCockpit
                       │
        ┌──────────────┼──────────────┐
        │              │              │
     HTTP            Auth          Session
        │              │              │
     Routes           PAM          Cookie
        │
   Middleware
        │
     Context
        │
   Controllers
        │
      HTML
```

Y, sobre todo, se han separado las responsabilidades:

```text
HTTP       → net/http
Routing    → routes
Auth       → PAM
Sessions   → session
Seguridad  → middleware
Request data → context
Lógica HTTP → controllers
HTML        → templates
```

---

# 17. Próximos pasos

La siguiente evolución natural del sistema sería:

1. Añadir `session.Delete()` para cerrar sesión.
2. Crear `/logout`.
3. Expirar sesiones.
4. Añadir protección contra sesiones inexistentes o manipuladas.
5. Mejorar el manejo de errores.
6. Separar las claves del contexto del paquete `middleware`.
7. Añadir autorización además de autenticación.

La diferencia será:

```text
Autenticación
    ↓
¿Quién eres?

Autorización
    ↓
¿Qué puedes hacer?
```

Esto será especialmente importante para GoCockpit porque determinadas operaciones del sistema, como administrar servicios, no deberían depender solamente de que el usuario haya iniciado sesión.
