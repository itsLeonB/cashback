# Expense Splitting & Debt Calculation Documentation

## Overview

The expense splitting system is designed to:

1. **Allocate item costs among participants** (share calculation)
2. **Compute participant share totals** (items + fees)
3. **Record debts as immutable transactions**

This system follows a **ledger-first, append-only model** with a strict lifecycle:

```text
Draft → Ready → Confirm (with optional dry-run preview)
```

---

## Core Principles

### 1. Snapshot Before Ledger

- All calculations are first computed and stored in:
  ```
  group_expense_participants
  ```
- This acts as the **source of truth per expense**

---

### 2. Immutable Debt Transactions

- Once an expense is confirmed:
  - Debt transactions are created
  - They **cannot be updated or deleted**
- Any correction must be done via **new compensating entries (future scope)**

---

### 3. Separation of Responsibilities

| Component            | Responsibility                       |
| -------------------- | ------------------------------------ |
| Allocation           | Who consumed what (per item)         |
| Fee Split            | How other fees are distributed       |
| Participant Snapshot | Final per-user share totals          |
| Debt Transactions    | Ledger for cross-expense aggregation |

---

### 4. Single Payer Model

- Each expense has **exactly one payer**
- All debts are directed toward that payer (or through a proxy — see section 4)

---

### 5. No Partial Payments

- The system **only tracks obligations**
- Settlement / repayment is out of scope

---

## 1. Share Calculation (Allocation Logic)

### Core Concept

Each expense item has:

- `totalAmount` (unit price × quantity)
- `participants[]`
- optional `weight` per participant

Supports:

- **Equal split**
- **Weighted split**

---

### Allocation Steps

#### 1. Validate Participants

- At least 1 participant required
- Weight rules:
  - All weights = 0 → equal split
  - All weights > 0 → weighted split
  - Mixed (some zero, some positive) → ❌ invalid
  - Any negative weight → ❌ invalid

---

#### 2. Determine Weights

```text
if weightSum == 0:
    weights = [1, 1, 1, ...]
else:
    weights = participant.weight
```

---

#### 3. Compute Unit Value

```text
unitValue = totalAmount / sum(weights)
```

---

#### 4. Allocate Per Participant

```text
allocatedAmount = weight * unitValue
allocatedAmount = round(allocatedAmount, 2)
```

---

#### 5. Handle Rounding Remainder

```text
remainder = totalAmount - allocatedSum
```

Assigned to:

- participant with highest weight
- tie-breaker: lowest ProfileID (UUID comparison)

---

#### 6. Final Validation

```text
sum(allocatedAmounts) == totalAmount
```

---

## 2. Other Fee Calculation

Other fees are split across all expense participants using one of two strategies:

### EQUAL_SPLIT

- Fee is divided equally among all participants
- Used for flat fees (e.g., table charge, booking fee)

```text
shareAmount = fee.amount / participantCount
```

### ITEMIZED_SPLIT

- Fee is distributed proportionally to each participant's item share
- Used for percentage-based fees (e.g., tax, service charge)

```text
rate = fee.amount / expense.itemsTotal
shareAmount = participant.shareAmount * rate
```

---

## 3. Expense Total Calculation

```text
itemsTotal = sum(item.unitPrice * item.quantity)
feesTotal  = sum(otherFees)
totalAmount = itemsTotal + feesTotal
```

Expense status transitions:

| Status      | Condition                                    |
| ----------- | -------------------------------------------- |
| `DRAFT`     | No items, or some items have no participants |
| `READY`     | All items have at least one participant      |
| `CONFIRMED` | Confirmed by creator; immutable              |

Status is recalculated automatically whenever items or their participants change.

---

## 4. Participant Share Calculation

After item allocation and fee splitting, each participant's total share is:

```text
shareAmount = sum(item allocations) + sum(fee shares)
```

This is stored in `group_expense_participants.share_amount`.

---

## 5. Proxy Model

A participant can have a **proxy** — another participant who pays on their behalf.

### Rules

- Proxy must be one of the expense participants
- Proxy cannot be the payer
- Proxy cannot itself have a proxy (no chaining)
- A participant cannot proxy themselves

### Debt Generation with Proxy

When participant P has proxy X, and payer is A, two debt transactions are created:

```text
P → X = shareAmount   (X covered P's share)
X → A = shareAmount   (X owes the payer for P's share)
```

Without a proxy, a single transaction is created:

```text
P → A = shareAmount
```

---

## 6. Debt Calculation

### Process

1. Identify payer
2. For each participant (excluding payer):
   - If participant has a proxy → create two-leg debt (see section 5)
   - Otherwise → create direct debt to payer
3. Participants with `shareAmount == 0` are skipped

### Debt Transaction Fields

```text
lenderProfileID
borrowerProfileID
amount
transferMethodID
groupExpenseID
description
```

### Description Format

| Case                     | Description                                                          |
| ------------------------ | -------------------------------------------------------------------- |
| Direct debt              | `Share for group expense: <description>`                             |
| Proxy covers participant | `Covered share for group expense: <description>`                     |
| Payer owed by proxy      | `Covered <participantName>'s share for group expense: <description>` |

---

### Example (no proxy)

| User      | Share |
| --------- | ----- |
| A (payer) | 30    |
| You       | 30    |
| B         | 40    |

Debts:

```text
You → A = 30
B → A = 40
```

### Example (with proxy)

| User      | Share | Proxy |
| --------- | ----- | ----- |
| A (payer) | 30    | —     |
| You       | 30    | —     |
| B         | 40    | You   |

Debts:

```text
You → A = 30          (You owes payer directly)
B → You = 40          (You covered B's share)
You → A = 40          (You owes payer for B's share)
```

---

## 7. Lifecycle

### DRAFT

- Expense can be freely edited (items, fees, participants)
- Status is `DRAFT` when any item has no participants
- No debts created

---

### READY

- All items have at least one participant assigned
- Status transitions to `READY` automatically
- Expense is eligible for confirmation

---

### Dry-Run (Preview)

- Call `ConfirmDraft` with `dryRun=true`
- System computes:
  - item allocations
  - fee splits
  - participant share totals
- Snapshot is written to `group_expense_participants`
- Returns a full `ExpenseConfirmationResponse` with per-participant breakdown (items, fees, totals)
- Expense status is **not** changed

---

### CONFIRMED

- Call `ConfirmDraft` with `dryRun=false`
- Requires:
  - Status is `READY`
  - At least one item
  - A payer is set
- Expense status becomes `CONFIRMED`
- Participant snapshot is written
- `ExpenseConfirmed` event is enqueued → debt transactions are created asynchronously
- No further modification allowed
- `Processed` flag is set after debt transactions are created (idempotency guard)

---

## 8. Confirmation Response

Both dry-run and confirmed calls return an `ExpenseConfirmationResponse`:

```text
id
description
totalAmount
payer
participants[]:
  profile
  proxyProfile (if applicable)
  items[]:
    id, name, baseAmount, shareRate, shareAmount
  itemsTotal
  fees[]:
    id, name, baseAmount, shareRate, shareAmount
  feesTotal
  total (= itemsTotal + feesTotal)
  hasProxy
```

---

## 9. Bill Upload & Parsing Flow (OCR)

The bill upload and parsing flow is an asynchronous, multi-stage process that leverages Google Vision for OCR and LLM for structured parsing.

For a detailed technical breakdown, including sequence diagrams and implementation details, see [Bill Upload Flow](file:///home/leon/itsLeonB/cashback/docs/bill-upload-flow.md).

### High-level Stages:

1. **Upload**: Image is uploaded (direct or via presigned URL) and stored in GCS.
2. **OCR**: An asynchronous worker extracts raw text from the image using Google Vision.
3. **Parsing**: Another worker sends the extracted text to an LLM to generate a structured `NewGroupExpenseRequest`.
4. **Update**: The expense draft is automatically updated with the parsed items and fees.

---

## 10. Source of Truth

### Per Expense

```text
group_expense_participants
```

- stores final computed share totals per participant
- used for UI, confirmation preview, and debt generation

---

### Cross Expense

```text
debt_transactions
```

- used for aggregating net balances between users

---

### Rule

```text
ALL debt_transactions must be derived ONLY from group_expense_participants
```

---

## 11. Net Balance Calculation

Across multiple expenses:

```text
net(A, B) = sum(A → B) - sum(B → A)
```

- No graph optimization
- No settlement logic
- Simple aggregation only

---

## 12. Design Decisions

### Deterministic Allocation

- Same input → same output

### Decimal Arithmetic

- Uses `shopspring/decimal` to prevent floating-point errors

### Immutable Ledger

- Ensures auditability
- Prevents accidental inconsistencies

### Snapshot-Based Calculation

- Avoids recomputation from ledger
- Ensures consistency between preview and confirmation

### Explicit Confirmation Step

- Prevents incorrect debt recording

### Idempotency Guard

- `GroupExpense.Processed` flag prevents duplicate debt creation if the confirmation event is replayed

---

## 13. Edge Cases

| Case                        | Behavior                                                           |
| --------------------------- | ------------------------------------------------------------------ |
| Negative weight             | ❌ rejected                                                        |
| Mixed zero/non-zero weights | ❌ rejected                                                        |
| No participants on item     | ❌ cannot confirm                                                  |
| No items                    | ❌ cannot confirm                                                  |
| No payer set                | ❌ cannot confirm                                                  |
| Already confirmed           | ❌ rejected                                                        |
| `shareAmount == 0`          | participant skipped in debt generation                             |
| Rounding remainder          | assigned to highest-weight participant (lowest UUID as tiebreaker) |

---

## 14. Mental Model

```text
1. Split item costs (allocation per item)
2. Split other fees (equal or itemized)
3. Sum per participant (snapshot)
4. Record obligations (immutable ledger)
```

---

## 15. Summary

### Share Calculation

```text
total → weights → allocation → rounding → validation
```

### Fee Calculation

```text
EQUAL_SPLIT: fee / participantCount
ITEMIZED_SPLIT: fee * (participantShare / itemsTotal)
```

### Debt Calculation

```text
shareAmount → direct: participant → payer
            → proxy:  participant → proxy → payer
```

### Architecture

```text
allocation → fee split → snapshot → immutable ledger
```
