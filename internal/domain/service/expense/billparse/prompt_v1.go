package billparse

import (
	"github.com/itsLeonB/cashback/internal/core/service/langfuse"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/ezutil/v2"
)

// BillParsePrompt is the single source of truth for everything that touches
// the "parse-bill" LLM prompt. The prompt version, variable contract, and
// response parser are all co-located here so they can only change together.
//
// To ship a new prompt version:
//  1. Create the new prompt version in Langfuse.
//  2. Add a new BillParsePromptV2 here with the updated version + parser.
//  3. Swap the reference in groupExpenseServiceImpl.
//  4. Delete the old one once fully rolled over.
type BillParsePrompt struct {
	// PromptName is the Langfuse prompt identifier.
	PromptName string
	// Version is pinned — never fetched as "latest".
	Version int
	// Variables documents the expected template variables.
	// This is both documentation and a guard (see Compile below).
	Variables []string
}

var ActiveBillParsePrompt = BillParsePrompt{
	PromptName: "parse-bill",
	Version:    1,
	Variables:  []string{"not_detected_bill_string", "text_to_parse"},
}

// NOTE: When updating the Langfuse prompt "parse-bill", ensure to include the following instruction:
// "All numeric values in the provided text have been pre-processed into canonical form
// (thousands separators removed, decimal separator is '.'). Do not reformat or re-interpret these numbers."

func (p BillParsePrompt) GetOptions() langfuse.GetPromptOptions {
	return langfuse.GetPromptOptions{Version: &p.Version}
}

func (p BillParsePrompt) CompileVars(notDetectedStr, textToParse string) map[string]any {
	return map[string]any{
		"not_detected_bill_string": notDetectedStr,
		"text_to_parse":            textToParse,
	}
}

func (p BillParsePrompt) ParseResponse(raw string) (dto.NewGroupExpenseRequest, error) {
	return ezutil.Unmarshal[dto.NewGroupExpenseRequest]([]byte(raw))
}
