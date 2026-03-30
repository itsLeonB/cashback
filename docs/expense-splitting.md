# Expense Splitting & Debt Calculation Documentation

## Overview

The expense splitting system is designed to:

1. **Allocate item costs among participants** (share calculation)
2. **Compute participant balances**
3. **Record debts as immutable transactions**

This system follows a **ledger-first, append-only model** with a strict lifecycle:

```text
Draft → Preview (dry-run) → Confirm → Immutable ledger
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

| Component | Responsibility |
|----------|----------------|
| Allocation | Who consumed what |
| Participant Snapshot | Final per-user balances |
| Debt Transactions | Ledger for cross-expense aggregation |

---

### 4. Single Payer Model

- Each expense has **exactly one payer**
- All debts are directed toward that payer

---

### 5. No Partial Payments

- The system **only tracks obligations**
- Settlement / repayment is out of scope

---

## 1. Share Calculation (Allocation Logic)

### Core Concept

Each expense item has:

- `totalAmount`
- `participants[]`
- optional `weight`

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
  - Mixed → ❌ invalid

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
- tie-breaker: lowest ProfileID

---

#### 6. Final Validation

```text
sum(allocatedAmounts) == totalAmount
```

---

## 2. Expense Total Calculation

```text
totalAmount = sum(items) + sum(fees)
```

Expense status:

- `Draft` → incomplete
- `Ready` → valid for confirmation

---

## 3. Participant Balance Calculation

After allocation, each participant has:

```text
allocatedAmount
paidAmount
```

Balance is:

```text
balance = paidAmount - allocatedAmount
```

---

### Interpretation

| Balance | Meaning |
|--------|--------|
| > 0    | Is owed money |
| < 0    | Owes money |
| = 0    | Settled |

---

## 4. Debt Calculation

### Model

- Single payer
- No proxy (in current logic)
- No partial payments

---

### Process

1. Identify payer
2. For each participant (excluding payer):

```text
debt = allocatedAmount
```

3. Create debt:

```text
participant → payer = allocatedAmount
```

---

### Example

| User | Paid | Share | Balance |
|------|------|-------|--------|
| A    | 100  | 30    | +70    |
| You  | 0    | 30    | -30    |
| B    | 0    | 40    | -40    |

Debts:

```text
You → A = 30
B → A = 40
```

---

## 5. Debt Transaction Model

Each debt is recorded as an **immutable ledger entry**:

```text
fromProfileID
toProfileID
amount
group_expense_id
description
```

---

### Key Characteristics

- Append-only
- No updates
- No deletes
- Created only after confirmation

---

### Description Format

Debt transactions include a standardized description:

```text
[DIRECT] Share for "<expense_name>"
```

Example:

```text
[DIRECT] Share for "Dinner"
```

---

## 6. Lifecycle

### Draft

- Expense can be freely edited
- No debts created

---

### Preview (Dry Run)

- System computes:
  - allocations
  - balances
  - resulting debts
- Stored in:
  ```
  group_expense_participants
  ```
- User must verify before confirmation

---

### Confirmed

- Expense becomes immutable
- Debt transactions are generated from snapshot
- No further modification allowed

---

## 7. Source of Truth

### Per Expense

```text
group_expense_participants
```

- stores final computed balances
- used for UI and reporting

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

## 8. Net Balance Calculation

Across multiple expenses:

```text
net(A, B) = sum(A → B) - sum(B → A)
```

---

### Properties

- No graph optimization
- No settlement logic
- Simple aggregation only

---

## 9. Design Decisions

### Deterministic Allocation

- Same input → same output

---

### Decimal Arithmetic

- Prevents floating point errors

---

### Immutable Ledger

- Ensures auditability
- Prevents accidental inconsistencies

---

### Snapshot-Based Calculation

- Avoids recomputation from ledger
- Ensures consistency

---

### Explicit Confirmation Step

- Prevents incorrect debt recording

---

## 10. Edge Cases

### Invalid Weights

- negative → ❌
- mixed zero/non-zero → ❌

---

### No Participants

- ❌ cannot allocate

---

### Negative Amount

- ❌ rejected

---

### Rounding Issues

- resolved via remainder assignment

---

## 11. Mental Model

The system operates as:

```text
1. Split cost (allocation)
2. Compute balances (snapshot)
3. Record obligations (ledger)
```

---

## 12. Summary

### Share Calculation

```text
total → weights → allocation → rounding → validation
```

---

### Debt Calculation

```text
allocatedAmount → participant → payer → ledger entry
```

---

### Architecture

```text
allocation → snapshot → immutable ledger
```

---

This design ensures:

- Accuracy
- Simplicity
- Auditability
- Deterministic behavior
