# Agents

## Verification

After making code changes, verify the implementation using:

```bash
make lint
make build-all
make test
```

Do not use `go build` or `go test` directly.

## Project Structure

```
cmd/
  http/       → HTTP server entrypoint
  worker/     → Background worker entrypoint
  job/        → One-off jobs (migrations, asset sync)
internal/
  appconstant/   → Application-wide constants and enums
  core/          → Framework/infra layer (config, logger, otel, services like cache/mail/queue)
  domain/
    dto/         → Request/response data transfer objects
    entity/      → Domain entities (DB models via go-crud)
    mapper/      → Entity ↔ DTO conversion functions
    message/     → Async message structs (for worker subscribers)
    repository/  → Repository interfaces
    service/     → Business logic (service interfaces + implementations)
  adapters/
    http/        → HTTP server, routes, handlers, middlewares
    repository/  → Repository implementations (GORM)
    core/        → Infrastructure service implementations
    db/          → Database setup (postgres)
    worker/      → Worker subscribers and schedulers
    job/         → Job implementations
  provider/      → Dependency injection / wiring
```

## Conventions

### Architecture

- Clean Architecture: domain layer has no dependency on adapters or framework.
- Service interfaces are defined in `internal/domain/service/services.go`.
- Repository interfaces are defined in `internal/domain/repository/`.
- All wiring happens in `internal/provider/` — constructors use plain dependency injection (no DI container).

### Naming

- Service implementations: `<name>ServiceImpl` struct, `New<Name>Service` constructor.
- Handlers: `<Name>Handler` struct with `Handle<Action>()` methods returning `gin.HandlerFunc`.
- DTOs: `<Action>Request` / `<Action>Response` in the `dto` package.
- Mappers: standalone functions in `mapper/`, named `<Entity>To<DTO>` or `<DTO>To<Entity>`.

### Error Handling

- Use `github.com/itsLeonB/ungerr` for typed errors (`UnauthorizedError`, `NotFoundError`, `Unknown`, `Wrap`).
- Never return raw `errors.New()` from domain services — always use `ungerr`.

### Testing

- Tests use `testify/assert`.
- Test files use `_test` package suffix (e.g., `package expense_test`).
- Use `github.com/vektra/mockery` for generating mocks.

### Dependencies (key libraries)

- HTTP framework: `gin-gonic/gin`
- ORM: `github.com/itsLeonB/go-crud` (wraps GORM with `BaseEntity`, `Transactor`)
- JWT: `github.com/itsLeonB/sekure`
- Validation: gin's `binding` tags on DTOs
- Decimal: `github.com/shopspring/decimal`
- UUID: `github.com/google/uuid`
- Config: `github.com/kelseyhightower/envconfig` (via `split_words` / `default` struct tags)
- Observability: OpenTelemetry (`internal/core/otel`)

### Configuration

- All config is loaded from environment variables (struct tags in `internal/core/config/`).
- `.env` is auto-loaded via `github.com/joho/godotenv/autoload`.
- See `.env.example` for required variables.

### Adding a New Feature

1. Define or extend the service interface in `services.go`.
2. Implement in `internal/domain/service/`.
3. Add DTOs in `dto/`, mappers in `mapper/` if needed.
4. Add handler in `internal/adapters/http/handler/`.
5. Register route in `internal/adapters/http/routes/`.
6. Wire dependencies in `internal/provider/service_provider.go`.
