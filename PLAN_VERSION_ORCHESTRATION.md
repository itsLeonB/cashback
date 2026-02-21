# Admin Endpoint Orchestration Analysis

## Plan Resource Orchestration

The admin endpoint for Plans follows this orchestration pattern:

### 1. **Handler Layer** (`internal/adapters/http/handler/admin/plan_handler.go`)
- Contains HTTP handlers for CRUD operations
- Methods: `HandleCreate`, `HandleGetList`, `HandleGetOne`, `HandleUpdate`, `HandleDelete`
- Uses `server.Handler` wrapper for consistent response handling
- Extracts path parameters using `server.GetRequiredPathParam`
- Binds JSON requests using `server.BindJSON`

### 2. **Routes Layer** (`internal/adapters/http/routes/admin_routes.go`)
- Registers routes under `/admin/v1/plans`
- Protected by `middlewares.AdminAuth` middleware
- Routes:
  - `POST /admin/v1/plans` - Create
  - `GET /admin/v1/plans` - List
  - `GET /admin/v1/plans/:planID` - Get One
  - `PUT /admin/v1/plans/:planID` - Update
  - `DELETE /admin/v1/plans/:planID` - Delete

### 3. **Handler Provider** (`internal/adapters/http/handler/admin/handlers.go`)
- Wires handlers with services via dependency injection
- `ProvideHandlers` function creates handler instances

### 4. **Service Layer** (`internal/domain/service/monetization/plan_service.go`)
- Interface: `PlanService` with CRUD methods
- Implementation: `planService` struct
- Uses `crud.Transactor` for transaction management
- Uses `crud.Repository` for data access
- Business logic includes validation and relationship checks

### 5. **Repository Layer** (`internal/provider/repository_provider.go`)
- Provides `crud.Repository[monetization.Plan]`
- Registered in `ProvideRepositories` function

### 6. **Service Provider** (`internal/provider/service_provider.go`)
- Wires service with repositories
- Registered in `ProvideServices` function

### 7. **DTOs** (`internal/domain/dto/monetization/plan_dto.go`)
- `NewPlanRequest` - Create request
- `PlanResponse` - Response structure
- `UpdatePlanRequest` - Update request

### 8. **Mapper** (`internal/domain/mapper/monetization/plan_mapper.go`)
- `PlanToResponse` - Converts entity to DTO

### 9. **Context Keys** (`internal/appconstant/context_keys.go`)
- `ContextPlanID` - Path parameter key

---

## PlanVersion Resource Orchestration (Generated)

Following the same pattern, PlanVersion admin endpoints have been created:

### Files Created/Modified:

1. **Handler**: `internal/adapters/http/handler/admin/plan_version_handler.go`
   - CRUD handlers for PlanVersion

2. **Routes**: `internal/adapters/http/routes/admin_routes.go` (modified)
   - Added `/admin/v1/plan-versions` routes

3. **Handler Provider**: `internal/adapters/http/handler/admin/handlers.go` (modified)
   - Added `PlanVersionHandler` to `Handlers` struct

4. **Service**: `internal/domain/service/monetization/plan_version_service.go`
   - `PlanVersionService` interface and implementation

5. **DTOs**: `internal/domain/dto/monetization/plan_dto.go` (modified)
   - Added `NewPlanVersionRequest`, `PlanVersionResponse`, `UpdatePlanVersionRequest`

6. **Mapper**: `internal/domain/mapper/monetization/plan_mapper.go` (modified)
   - Added `PlanVersionToResponse` function

7. **Service Provider**: `internal/provider/service_provider.go` (modified)
   - Added `PlanVersion` service to `Services` struct
   - Wired in `ProvideServices` function

8. **Context Keys**: `internal/appconstant/context_keys.go` (modified)
   - Added `ContextPlanVersionID` constant

### Routes Available:
- `POST /admin/v1/plan-versions` - Create
- `GET /admin/v1/plan-versions` - List
- `GET /admin/v1/plan-versions/:planVersionID` - Get One
- `PUT /admin/v1/plan-versions/:planVersionID` - Update
- `DELETE /admin/v1/plan-versions/:planVersionID` - Delete

### Key Features:
- Full CRUD operations
- Transaction support for data consistency
- Validation via binding tags
- Proper error handling with `ungerr` package
- Repository pattern with `go-crud` library
- Consistent response structure with DTOs
