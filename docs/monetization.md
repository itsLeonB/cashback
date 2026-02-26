# Monetization & Subscription System

This document serves as the single source of truth for the monetization system, covering the architecture, domain models, payment flows, and the concurrency-safe implementation logic.

---

## 1. High-Level Architecture

The monetization system manages user subscriptions and payment processing through a set of core components and external integrations.

### System Actors

- **Frontend**: Initiates purchases and renders the Midtrans Snap payment UI.
- **Backend (Go REST API)**: Manages plans, subscriptions, and payments. It implements strict validation and transactional integrity.
- **Midtrans**: The payment gateway for financial transactions.
- **Database (PostgreSQL)**: Persists all entities with row-level locking and unique constraints for data safety.

### Sync vs. Async Logic

- **Synchronous**: Purchase initialization and payment intent creation.
- **Asynchronous**: Payment settlement via Midtrans webhooks. Subscription state mutation occurs **atomically** within the webhook transaction (not in background workers).

---

## 2. Domain Models & States

### Plan & PlanVersion

- **Plan**: A logical product (e.g., "Pro Plan").
- **PlanVersion**: A specific pricing and limit snapshot. Active subscriptions are tied to a `PlanVersion` to prevent price changes from affecting existing users until renewal.

### Subscription

Tracks user access and billing cycles.

- **States**: `incomplete_payment`, `active`, `past_due_payment`, `canceled`.
- **Fields**: `CurrentPeriodStart`, `CurrentPeriodEnd`, `ProfileID`, `AutoRenew`.

### Payment

Tracks individual transaction attempts.

- **States**: `pending`, `processing`, `paid`, `canceled`, `error`, `expired`.
- **Fields**: `Amount`, `GatewayTransactionID`, `StartsAt`, `EndsAt`, `ExpiredAt`.

---

## 3. Core Payment Flows

### Flow 1: Purchase Initialization (New Subscription)

**Endpoint**: `POST /plans/:plan_id/versions/:plan_version_id/subscriptions`

- Reuses an existing `incomplete_payment` subscription if one exists for the plan version.
- Creates a `pending` payment intent.
- Returns a Midtrans Snap token.

### Flow 2: Make Payment (Retry/Renewal)

**Endpoint**: `POST /subscriptions/:subscription_id`

- **Idempotency**: Locks the subscription row (`FOR UPDATE`). Checks for existing `pending`/`processing` payments.
- **Intent Reuse**: If a valid (non-expired) intent exists, the backend returns the existing Snap token.
- **Expiration**: If a previous intent is expired, it is transitioned to `expired` status to free up the unique constraint slot for a new attempt.
- **Constraint**: Database-level unique index ensures only **one** incomplete payment intent can exist per subscription at a time.

### Flow 3: Midtrans Webhook Notification (Settlement)

**Endpoint**: `POST /payments/midtrans/notifications`
This flow is **atomic** and strictly transactional:

1. **Locking**: Locks the **Subscription** row first, then the **Payment** row (prevents deadlocks).
2. **Authoritative Status**: Calls Midtrans `CheckTransaction` API to verify the payload.
3. **Period Calculation**:
   - `startsAt` is anchored to `subscription.CurrentPeriodEnd` for active subscriptions (safe early renewal).
   - `startsAt` defaults to `time.Now()` for lapsed/incomplete subscriptions.
4. **Atomic Mutation**: The Payment status and Subscription state/periods are updated in the **same transaction**. Background workers are only used for non-financial side effects (logs, notifications).

---

## 4. Concurrency & Safety Guarantees

### Database-Level Protection

- **Unique Partial Index**: Ensures exactly one active intent per subscription.
- **Row-Level Locking**: `SELECT FOR UPDATE` on subscriptions serializes payment and settlement attempts per user.

### Race Condition Validation

| Scenario                           | System Action                                                                 |
| :--------------------------------- | :---------------------------------------------------------------------------- |
| **Simultaneous MakePayment calls** | First acquires lock, second waits and returns the first's token (Idempotent). |
| **Simultaneous Webhooks**          | First settles, second sees `paid` status and exits early.                     |
| **Early Renewal**                  | New period is appended to `CurrentPeriodEnd`. User loses no prepaid time.     |
| **Double Extension Attack**        | Prevented by atomic transaction and terminal state checks.                    |

---

## 5. Admin API Orchestration

The system follows a consistent orchestration pattern for Plan, PlanVersion, and Subscription management.

### Endpoint Mapping

| Resource          | Base Path                 | Key Methods                 |
| :---------------- | :------------------------ | :-------------------------- |
| **Plans**         | `/admin/v1/plans`         | CRUD, Feature toggle        |
| **Versions**      | `/admin/v1/plan-versions` | CRUD, Pricing configuration |
| **Subscriptions** | `/admin/v1/subscriptions` | CRUD, Relationship updates  |

### Performance Features

- Null-safe handling for termination dates (`EndsAt`, `CanceledAt`).
- Transaction-backed CRUD operations via `crud.Transactor`.
- `X-Total-Count` headers for paginated admin lists.

---

## 6. Security & Observability

- **Webhook Verification**: SHA512 signature hashing against `ServerKey`.
- **Price Integrity**: Amounts are strictly derived from `PlanVersion` records, never trusted from client input.
- **Correlation**: `Payment.ID` links application logs, Gateway Order IDs, and task metadata.
- **Secret Hygiene**: Sensitive tokens and keys are never logged.
