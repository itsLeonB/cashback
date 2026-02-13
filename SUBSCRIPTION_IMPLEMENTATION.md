# Subscription Admin Endpoint Implementation

## Summary

Implemented complete admin endpoint orchestration for the Subscription resource following the established pattern from Plan and PlanVersion resources.

## Files Created

### 1. Handler
- **Path**: `internal/adapters/http/handler/admin/subscription_handler.go`
- **Content**: CRUD handlers (Create, GetList, GetOne, Update, Delete)

### 2. Service
- **Path**: `internal/domain/service/monetization/subscription_service.go`
- **Content**: Business logic with transaction support

### 3. DTOs
- **Path**: `internal/domain/dto/monetization/subscription_dto.go`
- **Content**: 
  - `NewSubscriptionRequest`
  - `SubscriptionResponse`
  - `UpdateSubscriptionRequest`

### 4. Mapper
- **Path**: `internal/domain/mapper/monetization/subscription_mapper.go`
- **Content**: `SubscriptionToResponse` function

## Files Modified

### 1. Handler Provider
- **Path**: `internal/adapters/http/handler/admin/handlers.go`
- **Change**: Added `SubscriptionHandler` to `Handlers` struct

### 2. Routes
- **Path**: `internal/adapters/http/routes/admin_routes.go`
- **Change**: Added `/admin/v1/subscriptions` route group

### 3. Service Provider
- **Path**: `internal/provider/service_provider.go`
- **Changes**:
  - Added `Subscription` field to `Services` struct
  - Wired service in `ProvideServices` function

### 4. Repository Provider
- **Path**: `internal/provider/repository_provider.go`
- **Changes**:
  - Added `Subscription` repository to `Repositories` struct
  - Wired repository in `ProvideRepositories` function

### 5. Context Keys
- **Path**: `internal/appconstant/context_keys.go`
- **Change**: Added `ContextSubscriptionID` constant

## API Endpoints

All endpoints are protected by `AdminAuth` middleware:

- `POST /admin/v1/subscriptions` - Create subscription
- `GET /admin/v1/subscriptions` - List all subscriptions
- `GET /admin/v1/subscriptions/:subscriptionID` - Get single subscription
- `PUT /admin/v1/subscriptions/:subscriptionID` - Update subscription
- `DELETE /admin/v1/subscriptions/:subscriptionID` - Delete subscription

## Features

- ✅ Full CRUD operations
- ✅ Transaction support for data consistency
- ✅ Validation via binding tags
- ✅ Proper error handling with `ungerr` package
- ✅ Repository pattern with `go-crud` library
- ✅ Consistent response structure with DTOs
- ✅ Null-safe handling for `EndsAt` and `CanceledAt` fields
- ✅ X-Total-Count header for list endpoint
