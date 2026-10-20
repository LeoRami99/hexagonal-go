# 🛑 Go Hexagonal API Boilerplate — Gin

Boilerplate profesional para construir APIs REST escalables, mantenibles y testeables en **Go**, utilizando **Arquitectura Hexagonal (Ports & Adapters)** y **Gin** como framework HTTP.

El objetivo de este proyecto es proporcionar una base sólida para aplicaciones backend reales, manteniendo una separación estricta entre:

* reglas de negocio;
* casos de uso;
* transporte HTTP;
* persistencia;
* infraestructura;
* servicios externos;
* configuración.

La arquitectura está diseñada para permitir que componentes como Gin, PostgreSQL, Redis, proveedores de almacenamiento o servicios externos puedan ser reemplazados sin modificar el núcleo de negocio de la aplicación.

> **Implementación ejecutable:** este repositorio incluye una referencia funcional con CRUD, autenticación, RBAC, middlewares, configuración, pruebas y artefactos de despliegue. Consulta [`TEMPLATE.md`](TEMPLATE.md) para ver qué está implementado, cómo probarlo y cuáles son los puntos de extensión deliberados.

---

# 🧠 Principios arquitectónicos

Este proyecto sigue principalmente:

* Hexagonal Architecture / Ports & Adapters
* Dependency Inversion Principle
* Separation of Concerns
* Dependency Injection
* Domain-oriented design
* Explicit dependencies
* Fail Fast
* Graceful Shutdown
* Twelve-Factor App
* Testability by design

La regla principal es:

> Las dependencias siempre deben apuntar hacia el núcleo de la aplicación.

El dominio no debe conocer detalles relacionados con:

* Gin;
* Ent;
* PostgreSQL;
* Redis;
* JWT;
* HTTP;
* variables de entorno;
* Docker;
* proveedores externos.

---

# 🏗️ Arquitectura

La aplicación se divide conceptualmente en cuatro áreas principales.

```text
┌───────────────────────────────────────┐
│            Infrastructure             │
│ Gin · Ent · PostgreSQL · Redis · JWT  │
│ HTTP · Storage · Email · External API │
├───────────────────────────────────────┤
│               Adapters                │
│ Handlers · Repositories · Providers   │
├───────────────────────────────────────┤
│              Application              │
│       Services / Use Cases / Ports    │
├───────────────────────────────────────┤
│                Domain                 │
│ Entities · Value Objects · Errors     │
│         Business Rules                │
└───────────────────────────────────────┘
```

## Domain

Contiene las reglas fundamentales del negocio.

Debe mantenerse como **Go puro** siempre que sea posible.

Ejemplos:

```text
internal/core/domain/
├── user.go
├── role.go
├── errors.go
└── value_objects.go
```

El dominio no debe contener:

```go
json:"..."
ent/schema metadata
binding:"..."
```

Por ejemplo:

```go
type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
}
```

No:

```go
type User struct {
	ID    string `json:"id"`
	Email string `json:"email" binding:"required"`
}
```

---

# 🔌 Ports

Los puertos representan contratos.

Permiten que el Core dependa de abstracciones y no de implementaciones concretas.

```text
internal/core/ports/
├── user_repository.go
├── user_service.go
├── token_provider.go
├── cache.go
└── mailer.go
```

## Input Ports

Definen las operaciones que la aplicación permite ejecutar.

Ejemplo:

```go
type UserService interface {
	Create(ctx context.Context, input CreateUserInput) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, id string, input UpdateUserInput) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}
```

## Output Ports

Representan dependencias necesarias por la aplicación.

Ejemplo:

```go
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}
```

El Core conoce `UserRepository`, pero no sabe si detrás existe:

* PostgreSQL;
* MySQL;
* MongoDB;
* memoria;
* API externa.

---

# ⚙️ Services / Use Cases

Los servicios implementan los casos de uso y coordinan las reglas de negocio.

```text
internal/core/services/
├── user_service.go
├── auth_service.go
└── user_service_test.go
```

Ejemplo:

```go
type userService struct {
	userRepository ports.UserRepository
}

func NewUserService(
	userRepository ports.UserRepository,
) ports.UserService {
	return &userService{
		userRepository: userRepository,
	}
}
```

Los servicios pueden:

* validar reglas de negocio;
* consultar repositorios;
* ejecutar operaciones transaccionales;
* utilizar servicios externos mediante puertos;
* generar eventos de aplicación;
* retornar errores de dominio.

No deben:

* responder HTTP;
* usar `gin.Context`;
* conocer códigos HTTP;
* ejecutar queries directamente;
* leer variables de entorno.

---

# 🔌 Adapters

Los adaptadores conectan el Core con tecnologías externas.

Existen principalmente dos tipos.

## Driving Adapters

Inician operaciones dentro de la aplicación.

Ejemplos:

* REST API;
* CLI;
* cron jobs;
* consumers;
* workers;
* WebSocket.

En este proyecto el principal driving adapter es Gin.

## Driven Adapters

Son dependencias utilizadas por la aplicación.

Ejemplos:

* PostgreSQL;
* Redis;
* S3 / MinIO;
* SMTP;
* APIs externas;
* brokers de mensajería.

---

# 📂 Estructura recomendada

```text
go-hexagonal-api/
│
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── ent/                 # Cliente generado; no editar manualmente
│   │   └── schema/          # Schemas que sí se editan
│   ├── modules/
│   │   ├── users/
│   │   │   ├── domain/
│   │   │   │   ├── user.go
│   │   │   │   └── errors.go
│   │   │   ├── application/
│   │   │   │   ├── ports.go
│   │   │   │   ├── service.go
│   │   │   │   └── service_test.go
│   │   │   ├── infrastructure/
│   │   │   │   ├── persistence/
│   │   │   │   │   ├── memory/
│   │   │   │   │   └── ent/
│   │   │   │   │       ├── mapper.go
│   │   │   │   │       └── user_repository.go
│   │   │   │   ├── security/
│   │   │   │   └── mail/
│   │   │   └── delivery/
│   │   │       └── http/
│   │   │           └── v1/
│   │   │               ├── controller.go
│   │   │               └── dto.go
│   │   │
│   │   └── auth/
│   │       ├── domain/
│   │       ├── application/
│   │       ├── infrastructure/
│   │       │   ├── cache/
│   │       │   └── security/
│   │       └── delivery/
│   │           └── http/
│   │               └── v1/
│   │
│   └── platform/
│       ├── config/
│       ├── database/
│       ├── http/
│       │   ├── middleware/
│       │   └── response/
│       ├── logger/
│       └── server/
│
├── migrations/            # SQL versionado (formato Atlas) generado con cmd/migrate
│   └── atlas.sum
│
├── docs/
│   └── swagger/
│
├── scripts/
│
├── tests/
│   └── integration/
│
├── .github/
│   └── workflows/
│       ├── test.yml
│       └── deploy.yml
│
├── .env.example
├── .gitignore
├── .golangci.yml
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## Generar módulos y schemas Ent

Para crear la estructura hexagonal de un módulo sin preparar cada directorio manualmente:

```bash
make module MODULE=products
```

El nombre del módulo debe estar en `snake_case` y el comando no sobrescribe módulos existentes. Se generan las capas `domain`, `application`, `infrastructure` y `delivery/http/v1`, junto con archivos Go mínimos listos para completar.

Para crear un nuevo schema de Ent:

```bash
make ent-schema NAME=Product
```

`NAME` debe usar `PascalCase`. Este comando crea `internal/ent/schema/product.go` y ejecuta automáticamente `make ent-generate` para actualizar el cliente tipado. Después debes editar el schema generado para definir sus campos, relaciones e índices.

Cuando solo modifiques un schema existente, regenera el cliente con:

```bash
make ent-generate
```

Con `APP_ENV=development`, `make run` sincroniza automáticamente esos schemas con PostgreSQL mediante `Schema.Create`. En producción la aplicación no modifica el esquema.

---

# 🚀 Stack recomendado

| Componente          | Tecnología                 |
| ------------------- | -------------------------- |
| Language            | Go                         |
| HTTP                | Gin                        |
| Database            | PostgreSQL                 |
| ORM                 | Ent                        |
| Cache               | Redis                      |
| Validation          | go-playground/validator    |
| Authentication      | JWT                        |
| Password hashing    | Argon2id o bcrypt          |
| Schema management   | Ent Schema                 |
| Logging             | slog / logger estructurado |
| Documentation       | OpenAPI / Swagger          |
| Testing             | testing + testify          |
| Mocks               | mockery                    |
| Integration Testing | Testcontainers             |
| Observability       | OpenTelemetry              |
| Linter              | golangci-lint              |
| Containers          | Docker / Docker Compose    |
| CI/CD               | GitHub Actions             |

Las tecnologías de infraestructura pueden ser reemplazadas sin modificar el dominio.

Los schemas de Ent viven en `internal/ent/schema`. Puedes crear uno con `make ent-schema NAME=Product`; después de modificarlo, regenera el cliente tipado con `make ent-generate`.

Los schemas de Ent son la fuente de verdad. En desarrollo la aplicación ejecuta `client.Schema.Create(ctx)`; en producción se aplican migraciones versionadas generadas con `cmd/migrate`.

---

# ⚙️ Configuración

Toda configuración debe cargarse durante el bootstrap de la aplicación.

```text
Environment
      ↓
config.Load()
      ↓
main.go
      ↓
Dependencies
```

El Core nunca debe ejecutar directamente:

```go
os.Getenv("DATABASE_URL")
```

Se recomienda centralizar configuración:

```go
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}
```

Ejemplo `.env.example`:

```env
APP_ENV=development
APP_PORT=8080
APP_NAME=go-hexagonal-api

DATABASE_URL=postgres://postgres:postgres@localhost:5432/app?sslmode=disable
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5
DATABASE_CONN_MAX_LIFETIME=30m

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

JWT_SECRET=change-me
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

LOG_LEVEL=debug
```

Nunca se deben versionar secretos reales.

---

# 🗄️ Base de datos

La conexión debe ser inicializada una única vez durante el bootstrap.

Debe configurarse apropiadamente el connection pool:

```go
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(30 * time.Minute)
sqlDB.SetConnMaxIdleTime(5 * time.Minute)
```

Los valores deben ajustarse según:

* capacidad de PostgreSQL;
* número de réplicas del backend;
* carga esperada;
* infraestructura disponible.

---

# 🗃️ Esquema y migraciones

Los schemas ubicados en `internal/ent/schema/` son la fuente de verdad. La estrategia depende de `APP_ENV`:

- `development`: la aplicación ejecuta `client.Schema.Create(ctx)` al arrancar y sincroniza automáticamente el esquema.
- `production`: la aplicación nunca modifica el esquema; el despliegue debe aplicar previamente las migraciones versionadas de `migrations/`.

Flujo de desarrollo:

```bash
# Editar internal/ent/schema/*.go y regenerar.
make ent-generate
make run
```

Flujo para preparar producción. Las migraciones se gestionan con `cmd/migrate`, una herramienta en Go que usa la librería de Atlas: no requiere instalar Atlas CLI, ni `sh`, ni Docker, y funciona igual en Linux, macOS y Windows. Cada target de `make` equivale a `go run ./cmd/migrate <comando>`, por lo que en Windows sin `make` puedes ejecutarlo directamente:

```bash
# Generar una migración comparando los schemas Ent con el estado anterior.
make migration-diff NAME=add_products

# Revisar los archivos SQL generados y versionarlos.

# Consultar y aplicar migraciones usando DATABASE_URL de .env.
make migrate-status
make migrate-apply

# Equivalente sin make (cualquier SO):
go run ./cmd/migrate diff add_products
go run ./cmd/migrate status
go run ./cmd/migrate apply
```

Si la base ya había sido creada automáticamente por Ent antes de adoptar las migraciones versionadas, registra una única vez la migración inicial como baseline, usando el prefijo numérico del archivo:

```bash
make migrate-baseline VERSION=20260925003527
```

El baseline no ejecuta ese SQL; registra que la base ya se encuentra en esa versión. En bases de producción nuevas y vacías no se usa baseline, sino `make migrate-apply`.

`cmd/migrate` lee `DATABASE_URL` del entorno o de `.env`. `migration-diff` reproduce `migrations/` en una base vacía y la compara con los schemas Ent generados (ejecuta antes `make ent-generate`): usa `DEV_DATABASE_URL` si está definida o, si no, crea una base temporal en el servidor de `DATABASE_URL` (requiere permiso `CREATEDB`) y la elimina al terminar. Cada migración se aplica en su propia transacción, salvo que el archivo declare `-- atlas:txmode none`. El historial se guarda en `atlas_schema_revisions`, igual que Atlas CLI, así que ambas herramientas son compatibles.

Si editas a mano un archivo SQL, recalcula `migrations/atlas.sum` con `make migrate-hash`; si no, `apply` y `status` fallarán por checksum.

Los cambios destructivos deben revisarse y probarse antes de aplicarlos. En producción, ejecuta `make migrate-apply` como paso de despliegue antes de iniciar la nueva versión de la API.

---

# 🔄 Transacciones

Las operaciones que involucren múltiples escrituras relacionadas deben ejecutarse dentro de una transacción.

Ejemplos:

```text
Crear usuario
    ↓
Crear perfil
    ↓
Asignar rol
    ↓
Registrar auditoría
```

Si una operación falla, toda la transacción debe hacer rollback.

El manejo transaccional debe abstraerse para evitar que los casos de uso dependan directamente de Ent.

---

# 📦 DTOs

Los DTOs funcionan como frontera entre infraestructura y dominio.

## Request DTO

```go
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
```

## Response DTO

```go
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
```

Nunca retornar directamente entidades generadas por Ent desde los handlers.

Evitar:

```go
ctx.JSON(http.StatusOK, databaseUser)
```

Preferir:

```text
HTTP Request
    ↓
Request DTO
    ↓
Domain / Use Case
    ↓
Domain Entity
    ↓
Response DTO
    ↓
HTTP Response
```

---

# ✅ Validación

Existen dos clases diferentes de validación.

## Validaciones HTTP

Ejemplos:

* campo requerido;
* formato de email;
* longitud;
* UUID válido;
* rango numérico.

Deben ejecutarse en el adapter HTTP.

## Validaciones de negocio

Ejemplos:

* usuario ya existe;
* cuenta suspendida;
* vacante cerrada;
* límite alcanzado;
* operación no permitida.

Deben ejecutarse en el Core.

---

# ⚠️ Manejo de errores

Los errores deben ser explícitos y predecibles.

Ejemplo:

```go
var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUnauthorized = errors.New("unauthorized")
)
```

El servicio retorna errores de dominio.

```go
return nil, domain.ErrUserNotFound
```

El HTTP adapter decide su representación:

```go
switch {
case errors.Is(err, domain.ErrUserNotFound):
	return 404

case errors.Is(err, domain.ErrEmailAlreadyExists):
	return 409

default:
	return 500
}
```

El dominio nunca debe retornar:

```go
gin.H{}
```

ni:

```go
http.StatusNotFound
```

---

# 📡 Respuesta estándar de API

Se recomienda mantener un contrato consistente.

Respuesta exitosa:

```json
{
  "data": {
    "id": "01JXYZ...",
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

Error:

```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User not found"
  },
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

Errores de validación:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request",
    "fields": {
      "email": "must be a valid email"
    }
  }
}
```

---

# 📄 Pagination, Filtering & Sorting

Los endpoints de listado deben soportar un mecanismo uniforme.

Ejemplo:

```http
GET /api/v1/users?page=1&limit=20&sort=created_at&order=desc
```

Filtros:

```http
GET /api/v1/users?status=active&role=admin
```

Búsqueda:

```http
GET /api/v1/users?search=leonardo
```

Se recomienda limitar `limit` para evitar consultas excesivamente grandes.

---

# 🔐 Autenticación y autorización

Una API base debería soportar:

```text
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
```

Flujo:

```text
Credentials
    ↓
AuthService
    ↓
Validate Password
    ↓
Access Token
    +
Refresh Token
```

Se recomienda utilizar:

* Access Tokens de corta duración;
* Refresh Tokens con mayor duración;
* rotación/revocación cuando corresponda;
* almacenamiento seguro de secretos;
* HTTPS obligatorio en producción.

---

# 👮 Authorization

Autenticación y autorización son responsabilidades diferentes.

```text
Authentication
¿Quién eres?

Authorization
¿Qué puedes hacer?
```

La aplicación puede implementar:

* RBAC;
* permisos;
* scopes;
* policies.

Ejemplo:

```text
ADMIN
 ├── user:create
 ├── user:update
 ├── user:delete
 └── user:read

USER
 └── user:read:self
```

Las reglas críticas de autorización deben protegerse también en el caso de uso y no depender exclusivamente del middleware HTTP.

---

# 🔑 Password Security

Las contraseñas nunca deben almacenarse directamente.

Debe existir un puerto:

```go
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}
```

La implementación puede utilizar:

* Argon2id;
* bcrypt.

Esto evita acoplar el Core a una implementación criptográfica específica.

---

# 🌐 API Versioning

Todos los endpoints públicos deben utilizar versionado.

```text
/api/v1/users
/api/v1/auth/login
/api/v1/roles
```

Esto facilita introducir cambios incompatibles en el futuro.

---

# 🩺 Health Checks

El servidor debe proporcionar endpoints operacionales.

```http
GET /health
GET /ready
```

`/health` verifica que el proceso se encuentre activo.

Ejemplo:

```json
{
  "status": "ok"
}
```

`/ready` comprueba si la aplicación está preparada para recibir tráfico.

Puede validar:

* PostgreSQL;
* Redis;
* dependencias críticas.

Son especialmente importantes en:

* Docker;
* Kubernetes;
* balanceadores;
* despliegues rolling.

---

# 🧾 Logging estructurado

Evitar:

```go
fmt.Println("error creating user")
```

Preferir logs estructurados:

```go
logger.Error(
	"failed to create user",
	"error", err,
	"user_id", userID,
	"request_id", requestID,
)
```

Cada request debería tener un `request_id` o `correlation_id`.

Ejemplo:

```text
request_id=01KXYZ
method=POST
path=/api/v1/users
status=201
duration=42ms
```

Nunca registrar:

* passwords;
* JWT completos;
* API keys;
* secretos;
* información sensible innecesaria.

---

# 📊 Observabilidad

Para producción se recomienda soportar los tres pilares principales de observabilidad:

```text
Logs
Metrics
Traces
```

Puede utilizarse OpenTelemetry.

Ejemplos de métricas:

```text
http_requests_total
http_request_duration_seconds
db_query_duration_seconds
active_requests
errors_total
```

Las trazas permiten seguir operaciones como:

```text
HTTP Request
   ↓
UserService
   ↓
UserRepository
   ↓
PostgreSQL
```

---

# 🛡️ Middlewares

Middlewares recomendados:

```text
middleware/
├── auth.go
├── cors.go
├── logger.go
├── recovery.go
├── rate_limit.go
├── request_id.go
└── security_headers.go
```

Orden conceptual:

```text
Request
   ↓
Request ID
   ↓
Recovery
   ↓
Logger
   ↓
CORS
   ↓
Rate Limiter
   ↓
Authentication
   ↓
Authorization
   ↓
Handler
```

---

# 🚦 Rate Limiting

Los endpoints sensibles deben poder limitar solicitudes.

Especialmente:

```text
/auth/login
/auth/reset-password
/auth/register
```

Puede implementarse mediante:

* memoria para aplicaciones simples;
* Redis para entornos distribuidos.

Ejemplo conceptual:

```text
100 requests / minute / IP
```

Los valores deben configurarse según el endpoint y el contexto de negocio.

---

# 💾 Redis / Cache

Redis debe considerarse infraestructura opcional.

Casos de uso:

* cache;
* rate limiting;
* sesiones;
* refresh tokens;
* distributed locks;
* idempotency;
* datos temporales.

Debe accederse mediante interfaces.

Ejemplo:

```go
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(
		ctx context.Context,
		key string,
		value []byte,
		ttl time.Duration,
	) error
	Delete(ctx context.Context, key string) error
}
```

---

# 🔁 Context

Todos los métodos relacionados con I/O deberían recibir:

```go
context.Context
```

Ejemplo:

```go
FindByID(ctx context.Context, id string)
```

Esto permite manejar:

* timeouts;
* cancelación;
* tracing;
* propagación del ciclo de vida de la petición.

Nunca sustituir `context.Context` por `*gin.Context` dentro del Core.

---

# ⏱️ Timeouts

El servidor debe definir timeouts explícitos.

Ejemplo:

```go
server := &http.Server{
	Addr:              ":" + cfg.App.Port,
	Handler:           router,
	ReadHeaderTimeout: 5 * time.Second,
	ReadTimeout:       15 * time.Second,
	WriteTimeout:      15 * time.Second,
	IdleTimeout:       60 * time.Second,
}
```

También deben existir timeouts para:

* PostgreSQL;
* Redis;
* llamadas HTTP externas;
* servicios de almacenamiento.

---

# 🛑 Graceful Shutdown

La aplicación debe finalizar correctamente ante señales como:

```text
SIGINT
SIGTERM
```

Proceso esperado:

```text
SIGTERM
   ↓
Stop accepting requests
   ↓
Wait active requests
   ↓
Close HTTP server
   ↓
Close DB
   ↓
Close Redis
   ↓
Exit
```

Esto evita cortar requests o transacciones durante despliegues.

---

# 📚 OpenAPI / Swagger

La API debe documentar:

* endpoints;
* parámetros;
* request bodies;
* response bodies;
* códigos HTTP;
* autenticación;
* errores.

Idealmente:

```text
/swagger/index.html
```

La documentación debe reflejar el contrato real de la API.

---

# 🗂️ Storage

Para archivos se recomienda abstraer el proveedor.

```go
type Storage interface {
	Upload(
		ctx context.Context,
		key string,
		data io.Reader,
		contentType string,
	) error

	Delete(ctx context.Context, key string) error

	URL(ctx context.Context, key string) (string, error)
}
```

Esto permite intercambiar:

```text
MinIO
AWS S3
Cloudflare R2
Google Cloud Storage
```

sin modificar los casos de uso.

---

# 📧 Servicios externos

Email, SMS, WhatsApp y APIs externas también deben implementarse como adapters.

Ejemplo:

```go
type Mailer interface {
	Send(
		ctx context.Context,
		message MailMessage,
	) error
}
```

El dominio no debe saber si el proveedor real es:

```text
SendGrid
SES
SMTP
Mailgun
```

---

# ♻️ Idempotency

Para operaciones críticas puede implementarse soporte de:

```http
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
```

Especialmente útil para:

* pagos;
* creación de órdenes;
* webhooks;
* procesos asíncronos;
* operaciones susceptibles de reintento.

---

# 🔗 Webhooks

Los webhooks deben:

1. validar la firma del proveedor;
2. validar timestamp cuando corresponda;
3. evitar procesamiento duplicado;
4. responder rápidamente;
5. ejecutar procesos pesados de forma asíncrona;
6. mantener auditoría del evento.

---

# 📬 Jobs y procesamiento asíncrono

Cuando una tarea no necesita ejecutarse durante el request HTTP, puede procesarse en segundo plano.

Ejemplos:

* emails;
* generación de reportes;
* procesamiento de archivos;
* notificaciones;
* sincronizaciones;
* importaciones masivas.

Arquitectura futura:

```text
HTTP
 ↓
Use Case
 ↓
Queue
 ↓
Worker
 ↓
External Provider
```

---

# 🧪 Testing

La arquitectura permite probar cada capa independientemente.

## Unit Tests

Principalmente para:

```text
internal/core/services/
internal/core/domain/
```

Los repositorios se reemplazan por mocks.

Ejemplo:

```text
UserService
    ↓
MockUserRepository
```

Comando:

```bash
go test ./... -v
```

---

# 🔬 Integration Tests

Para repositories se recomienda utilizar una instancia real de PostgreSQL mediante Testcontainers.

Ejemplo:

```text
Test
 ↓
Testcontainers
 ↓
PostgreSQL Docker
 ↓
Repository
 ↓
Assertions
```

Esto ofrece mayor confianza que utilizar SQLite para validar comportamiento específico de PostgreSQL.

---

# 🌐 HTTP Tests

Gin puede probarse utilizando:

```go
net/http/httptest
```

Ejemplo:

```text
HTTP Request
    ↓
Gin Router
    ↓
Handler
    ↓
Mock Service
    ↓
HTTP Response
```

Debe verificarse:

* HTTP status;
* body;
* headers;
* validaciones;
* autenticación;
* manejo de errores.

---

# 📈 Coverage

Generar coverage:

```bash
go test ./... -coverprofile=coverage.out
```

Visualizar:

```bash
go tool cover -html=coverage.out
```

La cobertura debe utilizarse como indicador de riesgo y no como objetivo aislado.

---

# 🧹 Static Analysis

Antes de realizar merge debe ejecutarse:

```bash
go fmt ./...
go vet ./...
go test ./...
```

También se recomienda:

```bash
golangci-lint run
```

---

# 🐳 Docker

El proyecto debe poder ejecutarse mediante Docker.

Ejemplo:

```bash
docker compose up -d
```

Servicios locales:

```text
API
PostgreSQL
Redis
```

Ejemplo conceptual:

```yaml
services:

  api:
    build: .
    ports:
      - "8080:8080"

  postgres:
    image: postgres
    environment:
      POSTGRES_DB: app
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres

  redis:
    image: redis
```

Para producción se recomienda utilizar un Dockerfile multi-stage.

---

# 🛠️ Makefile

Se recomienda proporcionar comandos estándar.

```bash
make run
make build
make test
make lint
make ent-generate
make migration-diff NAME=add_products
make migrate-baseline VERSION=20260925003527
make migrate-status
make migrate-apply
make migrate-hash
make docker-up
make docker-down
```

Ejemplo:

```makefile
run:
	go run ./cmd/api

test:
	go test ./... -v

lint:
	golangci-lint run

build:
	go build -o bin/api ./cmd/api
```

---

# 🚀 Instalación

## 1. Requisitos

* Go 1.26 o versión compatible declarada en `go.mod`
* Git
* Docker y Docker Compose
* PostgreSQL
* Redis, si está habilitado

Las migraciones no requieren instalar nada adicional: se ejecutan con Go (`go run ./cmd/migrate`).

---

## 2. Clonar

```bash
git clone https://github.com/tu-organizacion/go-hexagonal-api.git
cd go-hexagonal-api
```

---

## 3. Variables de entorno

```bash
cp .env.example .env
```

Completar las variables requeridas.

> `make run` y los comandos de migración (`make migrate-*` o `go run ./cmd/migrate`) cargan automáticamente `.env`. Si ejecutas `go run` directamente, debes exportar las variables antes:
>
> ```bash
> set -a
> . ./.env
> set +a
> ```
>
> Otra opción es proporcionar únicamente las variables necesarias al ejecutar el comando, por ejemplo: `APP_PORT=8081 go run ./cmd/api`.

---

## 4. Dependencias

```bash
go mod tidy
```

---

## 5. Infraestructura local

```bash
docker compose up -d postgres redis
```

---

## 6. Generar el cliente Ent

Solo es necesario después de modificar un schema:

```bash
make ent-generate
```

---

## 7. Ejecutar

```bash
make run
```

o:

```bash
go run ./cmd/api
```

Por defecto:

```text
http://localhost:8080
```

Se recomienda ejecutar el paquete completo con `go run ./cmd/api`, y no solamente el archivo `cmd/api/main.go`. De esta forma Go incluye correctamente todos los archivos que puedan añadirse posteriormente al paquete `main`.

Verificar que la API esté disponible:

```bash
curl http://localhost:8080/health
```

Respuesta esperada:

```json
{
  "status": "ok"
}
```

### Puerto ocupado

Si el arranque finaliza con el siguiente error:

```text
listen tcp :8080: bind: address already in use
```

significa que otro proceso ya está escuchando en el puerto `8080`. Se puede identificar con:

```bash
lsof -nP -iTCP:8080 -sTCP:LISTEN
```

Después de comprobar que es seguro detenerlo, usar el PID mostrado por el comando:

```bash
kill <PID>
```

Como alternativa, iniciar esta API en otro puerto sin detener el proceso existente:

```bash
APP_PORT=8081 go run ./cmd/api
```

En ese caso, el health check estará disponible en:

```bash
curl http://localhost:8081/health
```

Los mensajes `[GIN-debug]` son normales en desarrollo. Al usar `APP_ENV=production`, el router se ejecuta en modo release y la configuración exige un `JWT_SECRET` de al menos 32 caracteres.

---

# 🏭 Build

```bash
go build -o bin/api ./cmd/api
```

Ejecutar:

```bash
./bin/api
```

---

# 🔄 CI/CD

Todo Pull Request debería ejecutar como mínimo:

```text
Checkout
   ↓
Setup Go
   ↓
Download dependencies
   ↓
Format Check
   ↓
go vet
   ↓
golangci-lint
   ↓
Unit Tests
   ↓
Integration Tests
   ↓
Build
   ↓
Security Checks
```

El deployment debe realizarse únicamente si las validaciones anteriores son exitosas.

---

# 🔒 Seguridad

La aplicación debe considerar como mínimo:

* HTTPS en producción;
* passwords hasheadas;
* secretos fuera del repositorio;
* JWT con expiración;
* refresh-token rotation cuando aplique;
* rate limiting;
* CORS configurado explícitamente;
* límites de tamaño del request;
* validación estricta de entrada;
* queries parametrizadas;
* protección contra mass assignment;
* timeouts HTTP;
* security headers;
* dependency scanning;
* logs sin credenciales;
* privilegios mínimos de base de datos.

Nunca almacenar secretos reales en:

```text
.env.example
Dockerfile
docker-compose.yml
GitHub
source code
```

---

# 🧩 Dependency Injection

`cmd/api/main.go` se encarga de iniciar y detener el servidor. `cmd/api/bootstrap.go` funciona como **Composition Root** y construye las dependencias.

`internal/platform/http/router.go` configura los middleware globales y agrupa `/api/v1`; cada módulo registra sus endpoints en `internal/modules/{auth,users,courses}/delivery/http/v1/routes.go`.

Conceptualmente:

```go
func main() {
	cfg := config.Load()

	logger := logger.New(cfg)

	db := database.NewPostgres(cfg.Database)

	userRepository := postgres.NewUserRepository(db)

	userService := services.NewUserService(
		userRepository,
	)

	userHandler := http.NewUserHandler(
		userService,
	)

	router := setupRouter(userHandler)

	server := setupServer(router)

	run(server)
}
```

Esto permite visualizar claramente:

```text
Infrastructure
     ↓
Repositories
     ↓
Services
     ↓
Handlers
     ↓
Router
     ↓
Server
```

No se recomienda utilizar variables globales para compartir dependencias.

---

# 📐 Convenciones

## Naming

Packages:

```text
user
repository
service
middleware
config
```

Evitar nombres:

```text
utils
helpers
common
misc
```

cuando terminan agrupando responsabilidades sin relación.

---

## Constructors

Utilizar funciones constructoras:

```go
func NewUserService(...) *UserService
```

```go
func NewUserRepository(...) *UserRepository
```

```go
func NewUserHandler(...) *UserHandler
```

---

## Interfaces

Las interfaces deben existir para representar una frontera arquitectónica real.

Evitar crear interfaces innecesarias únicamente "porque es arquitectura hexagonal".

Preferir interfaces pequeñas y específicas.

---

# 📌 Flujo completo de una petición

Ejemplo:

```text
POST /api/v1/users
        │
        ▼
┌────────────────────┐
│ Gin Router         │
└─────────┬──────────┘
          ▼
┌────────────────────┐
│ Middleware         │
│ Auth / Logger      │
└─────────┬──────────┘
          ▼
┌────────────────────┐
│ UserHandler        │
│ HTTP Adapter       │
└─────────┬──────────┘
          │
          │ Request DTO
          ▼
┌────────────────────┐
│ UserService        │
│ Use Case           │
└─────────┬──────────┘
          │
          │ Domain User
          ▼
┌────────────────────┐
│ UserRepository     │
│ Port               │
└─────────┬──────────┘
          ▼
┌────────────────────┐
│ PostgresRepository │
│ Adapter            │
└─────────┬──────────┘
          ▼
┌────────────────────┐
│ PostgreSQL         │
└────────────────────┘
```

El resultado vuelve recorriendo la arquitectura en sentido inverso.

---

# ✅ Checklist mínimo para producción

Antes de considerar el boilerplate listo para producción debería contar con:

* [ ] Arquitectura Hexagonal
* [ ] Configuración centralizada
* [ ] Gin Router
* [ ] PostgreSQL
* [ ] Connection Pool
* [ ] Repository Pattern
* [ ] Estrategia de cambios destructivos y respaldo del esquema
* [ ] Transactions
* [ ] DTO Request / Response
* [ ] Validation
* [ ] Domain Errors
* [ ] Error Handler estándar
* [ ] Pagination
* [ ] Filtering
* [ ] Sorting
* [ ] Authentication
* [ ] Authorization
* [ ] Password Hashing
* [ ] Access / Refresh Tokens
* [ ] CORS
* [ ] Rate Limiting
* [ ] Request ID
* [ ] Structured Logging
* [ ] Recovery Middleware
* [ ] Health Check
* [ ] Readiness Check
* [ ] Graceful Shutdown
* [ ] HTTP Timeouts
* [ ] Database Timeouts
* [ ] Redis opcional
* [ ] OpenAPI / Swagger
* [ ] Unit Tests
* [ ] Integration Tests
* [ ] HTTP Tests
* [ ] Testcontainers
* [ ] Mocks
* [ ] Coverage
* [ ] golangci-lint
* [ ] Dockerfile
* [ ] Docker Compose
* [ ] Makefile
* [ ] CI/CD
* [ ] Observabilidad
* [ ] Métricas
* [ ] Tracing
* [ ] Seguridad de secretos

---

# 🚫 Qué no debe hacer el Core

El Core no debe importar directamente:

```text
github.com/gin-gonic/gin
entgo.io/ent
github.com/redis/go-redis
net/http
```

Tampoco debe:

```go
os.Getenv(...)
```

ni conocer:

```text
HTTP Status
JSON
SQL
Redis Keys
JWT implementation
Docker
Environment Variables
```

Su responsabilidad es representar:

```text
Business Rules
+
Use Cases
+
Contracts
```

---

# 🎯 Objetivo del boilerplate

La intención de esta plantilla no es simplemente organizar archivos.

Busca permitir que una aplicación evolucione desde:

```text
MVP
```

hacia:

```text
Production API
        ↓
Multiple Modules
        ↓
Workers / Queues
        ↓
Distributed Infrastructure
        ↓
Microservices cuando realmente sean necesarios
```

sin tener que reescribir el núcleo del negocio.

La arquitectura debe permitir reemplazar:

```text
Gin
PostgreSQL
Ent
Redis
JWT Provider
Storage Provider
Email Provider
```

con un impacto mínimo sobre los casos de uso.

---

# 🤝 Contribución

1. Crear una rama:

```bash
git checkout -b feature/nueva-funcionalidad
```

2. Implementar los cambios.

3. Ejecutar:

```bash
make lint
make test
```

4. Crear un Pull Request.

Los cambios deben mantener las reglas de dependencia y separación de responsabilidades establecidas por la arquitectura.

---

# 📄 Licencia

Este proyecto se distribuye bajo la licencia **MIT**.
