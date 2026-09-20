# Autenticación con PAM

## Objetivo

GoCockpit utilizará los usuarios existentes del sistema Linux para iniciar sesión.

No se crearán usuarios propios ni se almacenarán contraseñas en GoCockpit.

## ¿Qué es PAM?

PAM significa **Pluggable Authentication Modules**.

Es el sistema de autenticación utilizado por Linux que permite a los programas comprobar las credenciales de los usuarios mediante una configuración común.

GoCockpit delega la autenticación en PAM.

## Funcionamiento

El proceso de autenticación es:

```text
GoCockpit
    ↓
PAM
    ↓
Linux
    ↓
Usuario y contraseña
```

GoCockpit proporciona el usuario y la contraseña a PAM y recibe el resultado de la autenticación.

## Usuarios

Se utilizarán los usuarios existentes en Linux.

Por ejemplo, si el sistema tiene un usuario:

```text
gts
```

ese mismo usuario podrá utilizarse para iniciar sesión en GoCockpit.

Si la contraseña del usuario cambia en Linux, GoCockpit utilizará automáticamente la nueva contraseña.

## Configuración

GoCockpit utiliza un servicio PAM propio:

```text
/etc/pam.d/gocockpit
```

Configuración inicial:

```text
auth include system-login
account include system-login
```

Esta configuración permite utilizar los mecanismos de autenticación del sistema.

## Implementación en Go

La comunicación con PAM se realiza mediante la librería:

```text
github.com/msteinert/pam/v2
```

La autenticación está separada en el paquete:

```text
internal/auth/
```

La función principal es:

```go
func Authenticate(username, password string) error
```

Devuelve `nil` cuando la autenticación es correcta y un error cuando falla.

## Seguridad

GoCockpit no accederá directamente a:

```text
/etc/shadow
```

ni almacenará las contraseñas de los usuarios.

La autenticación se delega en PAM.

Además, la autenticación y la sesión web son conceptos diferentes:

* **PAM:** comprueba las credenciales.
* **Sesión de GoCockpit:** mantiene al usuario conectado después de autenticarse.

## Prueba

Antes de integrarlo con HTTP se realizó una prueba directa desde Go:

```text
Usuario: usuario_linux
Contraseña: ********
✅ Autenticación correcta
```

También se comprobó el comportamiento con una contraseña incorrecta:

```text
Usuario: usuario_linux
Contraseña: ********
❌ Autenticación fallida
```

## Decisión

Se utilizará PAM para la autenticación de GoCockpit porque permite reutilizar los usuarios del sistema Linux sin implementar un sistema independiente de contraseñas.
