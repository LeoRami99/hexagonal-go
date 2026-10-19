# Guía de Documentación OpenAPI / Swagger en Go

Esta guía describe el estándar y las buenas prácticas para documentar la API REST utilizando **swaggo/swag** bajo la arquitectura hexagonal del proyecto.

---

## 1. Reglas Arquitectónicas

De acuerdo a las reglas del proyecto (`.agent/rules/project-contex.md`):

1. **Ubicación exclusiva**: Las anotaciones de Swagger pertenecen **únicamente a la capa de Delivery** (`internal/modules/<modulo>/delivery/http/v1/`).
2. **Dominio puro**: Las entidades de `domain/` y los casos de uso de `application/` **nunca** deben contener anotaciones `@Param`, `@Success`, `@Router`, ni tags de transporte (`json`, `binding`, `example`).
3. **Centralización en Platform**: La configuración global de Swagger (`@title`, `@version`, `@host`, `@BasePath`, `@securityDefinitions`) y el handler de la UI (`/swagger/*any`) residen en `internal/platform/http/router.go`.

---

## 2. Configuración Global (`router.go`)

En `internal/platform/http/router.go` se declaran los metadatos generales de la API y el esquema de seguridad:

```go
// @title                      Campus API
// @version                    1.0
// @description                API REST construida con Go, Gin y Arquitectura Hexagonal.

// @contact.name              API Support
// @contact.email             soporte@empresa.com

// @host                      localhost:8080
// @BasePath                  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Escribe "Bearer " seguido de tu token JWT
```

> **Nota:** Debido a que `@BasePath` es `/api/v1`, las rutas en las anotaciones `@Router` deben ser relativas a este prefijo (por ejemplo `/users` en lugar de `/api/v1/users`).

---

## 3. Estructura de Anotaciones en Controladores

Cada handler del controlador HTTP debe tener su bloque de comentarios antes de la función:

| Etiqueta | Descripción | Ejemplo |
| :--- | :--- | :--- |
| `@Summary` | Descripción breve (una línea) | `Obtener usuario por ID` |
| `@Description` | Descripción detallada de lo que hace el endpoint | `Busca y retorna los datos del usuario especificado por su identificador único.` |
| `@Tags` | Módulo o agrupación en la UI | `users`, `auth` |
| `@Accept` | Formato recibido en el request | `json` |
| `@Produce` | Formato entregado en el response | `json` |
| `@Security` | Mecanismo de autenticación requerido | `BearerAuth` (omitir en endpoints públicos) |
| `@Param` | Parámetro de URL, query o cuerpo | Ver sección de parámetros |
| `@Success` | Código HTTP exitoso y tipo retornado | `200 {object} response.Envelope{data=UserResponse}` |
| `@Failure` | Códigos HTTP de error y tipo de error | `400 {object} response.ErrorResponse "Datos inválidos"` |
| `@Router` | Ruta relativa y método HTTP entre corchetes | `/users/{id} [get]` |

---

## 4. Sintaxis de `@Param`

Formato:
```text
@Param <nombre> <ubicación> <tipo> <requerido> "<descripción>" [<extras>]
```

* **Parámetro en ruta (Path)**:
  ```go
  // @Param id path string true "ID único del recurso (ULID o UUID)"
  ```
* **Parámetro en Query (Query string)**:
  ```go
  // @Param page  query int    false "Número de página" default(1)
  // @Param limit query int    false "Cantidad de registros por página" default(20)
  // @Param role  query string false "Filtrar por rol" Enums(admin, student, instructor)
  ```
* **Cuerpo de la petición (JSON Body)**:
  ```go
  // @Param request body CreateUserRequest true "Payload de creación de usuario"
  ```

---

## 5. Respuestas con Envoltorio Estándar (`Envelope`)

El proyecto utiliza un formato de respuesta estándar definido en `internal/platform/http/response/`:
* `response.Envelope`: `{ "data": ..., "meta": ... }`
* `response.ErrorResponse`: `{ "error": { "code": "...", "message": "...", "fields": ... }, "request_id": "..." }`

### ¿Cómo documentar la respuesta en Swagger?

* **Objeto único**:
  ```go
  // @Success 200 {object} response.Envelope{data=UserResponse}
  ```
* **Listas paginadas**:
  ```go
  // @Success 200 {object} response.Envelope{data=[]UserResponse,meta=object}
  ```
* **Respuestas sin contenido (204 No Content)**:
  ```go
  // @Success 204 "Sin contenido"
  ```
* **Errores**:
  ```go
  // @Failure 400 {object} response.ErrorResponse "Petición inválida o datos malformados"
  // @Failure 401 {object} response.ErrorResponse "Token no proporcionado o inválido"
  // @Failure 403 {object} response.ErrorResponse "Permisos insuficientes"
  // @Failure 404 {object} response.ErrorResponse "Recurso no encontrado"
  // @Failure 500 {object} response.ErrorResponse "Error interno del servidor"
  ```

---

## 6. Documentación de DTOs (`dto.go`)

En los structs de entrada (`Request`) y salida (`Response`) en `delivery/http/v1/dto.go`, añade etiquetas `example` y comentarios explicativos:

```go
package v1

// CreateUserRequest define el payload para registrar un nuevo usuario.
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required" example:"Juan Pérez"`
	Email    string `json:"email" binding:"required,email" example:"juan.perez@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"Secret123!"`
	Role     string `json:"role" binding:"required,oneof=admin student instructor" example:"student"`
}

// UserResponse representa los datos públicos del usuario retornados al cliente.
type UserResponse struct {
	ID        string `json:"id" example:"01J8F5XYZ1234567890ABCDEFG"`
	Name      string `json:"name" example:"Juan Pérez"`
	Email     string `json:"email" example:"juan.perez@example.com"`
	Role      string `json:"role" example:"student"`
	Active    bool   `json:"active" example:"true"`
	CreatedAt string `json:"created_at" example:"2026-09-28T12:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2026-09-28T12:00:00Z"`
}
```

---

## 7. Ejemplos Prácticos Completos

### Ejemplo A: Endpoint Público (Login)

```go
// Login godoc
// @Summary      Iniciar sesión
// @Description  Autentica credenciales de usuario y genera un par de tokens JWT (Access y Refresh).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     LoginRequest true "Credenciales de acceso"
// @Success      200     {object} response.Envelope{data=TokenResponse}
// @Failure      400     {object} response.ErrorResponse "Formato de request inválido"
// @Failure      401     {object} response.ErrorResponse "Credenciales inválidas"
// @Router       /auth/login [post]
func (h *Controller) Login(c *gin.Context) {
    // ...
}
```

### Ejemplo B: Endpoint Protegido con Parámetro en Path

```go
// Get godoc
// @Summary      Obtener usuario
// @Description  Retorna la información de perfil de un usuario dado su ID.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path     string true "Identificador del usuario"
// @Success      200 {object} response.Envelope{data=UserResponse}
// @Failure      401 {object} response.ErrorResponse "No autenticado"
// @Failure      404 {object} response.ErrorResponse "Usuario no encontrado"
// @Router       /users/{id} [get]
func (h *Controller) Get(c *gin.Context) {
    // ...
}
```

### Ejemplo C: Endpoint Protegido con Filtros y Paginación

```go
// List godoc
// @Summary      Listar usuarios
// @Description  Retorna una lista paginada de usuarios con soporte para búsqueda y filtrado.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page   query    int    false "Número de página" default(1)
// @Param        limit  query    int    false "Límite por página" default(20)
// @Param        search query    string false "Término de búsqueda por nombre o email"
// @Param        role   query    string false "Filtrar por rol" Enums(admin, student, instructor)
// @Param        active query    bool   false "Filtrar por estado activo"
// @Success      200    {object} response.Envelope{data=[]UserResponse,meta=object}
// @Failure      401    {object} response.ErrorResponse "No autenticado"
// @Failure      403    {object} response.ErrorResponse "Requiere rol de Administrador"
// @Router       /users [get]
func (h *Controller) List(c *gin.Context) {
    // ...
}
```

---

## 8. Ciclo de Trabajo y Regeneración

Cada vez que agregues un nuevo endpoint o modifiques los structs DTO:

1. Agrega o actualiza las anotaciones en el controlador de Delivery correspondiente.
2. Ejecuta el comando en consola:
   ```bash
   make swagger
   ```
3. Inicia la aplicación (`make run` o `docker compose up`) y verifica los cambios en el navegador:
   ```text
   http://localhost:8080/swagger/index.html
   ```
4. Asegúrate de que los tests sigan pasando:
   ```bash
   make test
   ```
