# Bill Upload & Parsing Flow

The bill upload system automates expense creation by extracting items and fees from receipt images using Google Vision (OCR) and OpenAI (LLM).

## Overview

The process is asynchronous and event-driven, involving the frontend, a Go backend API, and a worker system.

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant API
    participant Storage
    participant Queue
    participant Worker
    participant OCR (Vision)
    participant LLM (OpenAI)

    User->>Frontend: Selects Image
    Frontend->>Frontend: Compress Image

    alt Traditional Upload
        Frontend->>API: POST /group-expenses/:id/bill
        API->>Storage: Upload Image
        API->>Queue: ExpenseBillUploaded
    else Presigned Upload (Feature Flag)
        Frontend->>API: POST /group-expenses/:id/bill/presigned
        API-->>Frontend: Presigned URL + BillID
        Frontend->>Storage: PUT Image
        Frontend->>API: POST /group-expenses/:id/bill/:billId/parse
        API->>Queue: ExpenseBillUploaded
    end

    Queue->>Worker: Handle ExpenseBillUploaded
    Worker->>OCR (Vision): Extract Text
    Worker->>Queue: ExpenseBillTextExtracted

    Queue->>Worker: Handle ExpenseBillTextExtracted
    Worker->>LLM (OpenAI): Parse Text to JSON
    Worker->>API: Update Expense Draft (Items/Fees)
    Worker->>Worker: Mark Bill as PARSED
```

---

## 1. Frontend Optimization

To improve reliability on unstable networks, images are compressed client-side before upload.

- **Utility**: `compressImageForOCR`
- **Logic**: Skips files < 800 KB; others are compressed while maintaining OCR-readable quality.
- **Library**: `browser-image-compression`

---

## 2. Upload Strategies

The system supports two upload strategies controlled by the `use_presigned_bill_upload` feature flag.

### Traditional Upload

1. Frontend sends a `multipart/form-data` request to the API.
2. API streams the file to storage and creates a DB record.
3. API enqueues `ExpenseBillUploaded`.

### Presigned Upload

1. Frontend requests a presigned URL.
2. API creates a DB record in `NOT_UPLOADED_BILL` status and returns a PUT URL.
3. Frontend uploads directly to Storage.
4. Frontend notifies API via `/parse` endpoint, which enqueues `ExpenseBillUploaded`.

---

## 3. Asynchronous Processing

### Stage 1: Text Extraction (OCR)

- **Message**: `ExpenseBillUploaded`
- **Handler**: `ExpenseBillService.ExtractBillText`
- **Logic**:
  - Retrieves image from Storage.
  - Sends to **Google Cloud Vision API**.
  - Stores raw text in `expense_bills.extracted_text`.
  - Updates status to `EXTRACTED`.
  - Enqueues `ExpenseBillTextExtracted`.

### Stage 2: Structured Parsing (LLM)

- **Message**: `ExpenseBillTextExtracted`
- **Handler**: `GroupExpenseService.ParseFromBillText`
- **Logic**:
  - Sends raw text to **OpenAI** with a specialized system prompt.
  - **Prompt Instruction**: Extract `totalAmount`, `subtotal`, `items`, and `otherFees` into a strict JSON schema.
  - Discards invalid or zero-amount items.
  - Updates the linked `GroupExpense` draft with the extracted data.
  - Updates status to `PARSED`.

---

## 4. Bill Lifecycle Status

| Status              | Description                                        |
| ------------------- | -------------------------------------------------- |
| `PENDING`           | Uploaded, waiting for OCR.                         |
| `EXTRACTED`         | OCR complete, waiting for LLM parsing.             |
| `PARSED`            | Successfully parsed into expense items.            |
| `FAILED`            | Critical error during OCR or parsing.              |
| `NOT_DETECTED`      | LLM could not find valid receipt data in the text. |
| `NOT_UPLOADED_BILL` | Presigned URL issued but upload not yet confirmed. |

---

## 5. Security & Constraints

- **File Limits**: Maximum file size is enforced at both API and Storage levels.
- **Allowed Extensions**: Only image MIME types (JPEG, PNG, WEBP) are accepted.
- **Subscription Limits**: Users may be restricted in the number of bills they can upload per month.
- **Ownership**: Only the expense creator can upload a bill to a draft.
