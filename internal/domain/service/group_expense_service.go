package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/llm"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/cashback/internal/domain/service/fee"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"github.com/shopspring/decimal"
)

type groupExpenseServiceImpl struct {
	friendshipService     FriendshipService
	expenseRepo           repository.GroupExpenseRepository
	transactor            crud.Transactor
	feeCalculatorRegistry map[expenses.FeeCalculationMethod]fee.FeeCalculator
	otherFeeRepository    repository.OtherFeeRepository
	billRepo              crud.Repository[expenses.ExpenseBill]
	llmService            llm.LLMService
	billSvc               ExpenseBillService
}

func NewGroupExpenseService(
	friendshipService FriendshipService,
	expenseRepo repository.GroupExpenseRepository,
	transactor crud.Transactor,
	feeCalculatorRegistry map[expenses.FeeCalculationMethod]fee.FeeCalculator,
	otherFeeRepository repository.OtherFeeRepository,
	billRepo crud.Repository[expenses.ExpenseBill],
	llmService llm.LLMService,
	billSvc ExpenseBillService,
) GroupExpenseService {
	return &groupExpenseServiceImpl{
		friendshipService,
		expenseRepo,
		transactor,
		feeCalculatorRegistry,
		otherFeeRepository,
		billRepo,
		llmService,
		billSvc,
	}
}

func (ges *groupExpenseServiceImpl) CreateDraft(ctx context.Context, userProfileID uuid.UUID, description string) (dto.GroupExpenseResponse, error) {
	newDraftExpense := expenses.GroupExpense{
		CreatorProfileID: userProfileID,
		Description:      description,
		Status:           expenses.DraftExpense,
	}

	insertedDraftExpense, err := ges.expenseRepo.Insert(ctx, newDraftExpense)
	if err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	return mapper.GroupExpenseToResponse(insertedDraftExpense, userProfileID, ""), nil
}

func (ges *groupExpenseServiceImpl) GetAllCreated(ctx context.Context, userProfileID uuid.UUID, status expenses.ExpenseStatus) ([]dto.GroupExpenseResponse, error) {
	spec := crud.Specification[expenses.GroupExpense]{}
	spec.Model.CreatorProfileID = userProfileID
	spec.PreloadRelations = []string{"Items", "OtherFees", "Participants", "Payer", "Creator"}
	spec.Model.Status = status

	groupExpenses, err := ges.expenseRepo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(groupExpenses, mapper.GroupExpenseSimpleMapper(userProfileID, "")), nil
}

func (ges *groupExpenseServiceImpl) GetDetails(ctx context.Context, id, userProfileID uuid.UUID) (dto.GroupExpenseResponse, error) {
	spec := crud.Specification[expenses.GroupExpense]{}
	spec.Model.ID = id
	spec.PreloadRelations = []string{
		"Items",
		"OtherFees",
		"Payer",
		"Creator",
		"Items.Participants",
		"Items.Participants.Profile",
		"Participants",
		"Participants.Profile",
		"Bill",
	}

	groupExpense, err := ges.getGroupExpense(ctx, spec)
	if err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	billURL, err := ges.billSvc.GetURL(ctx, groupExpense.Bill.ImageName)
	if err != nil {
		logger.Errorf("error retrieving bill image URL: %v", err)
	}

	return mapper.GroupExpenseToResponse(groupExpense, userProfileID, billURL), nil
}

func (ges *groupExpenseServiceImpl) ConfirmDraft(ctx context.Context, id, profileID uuid.UUID, dryRun bool) (dto.GroupExpenseResponse, error) {
	var response dto.GroupExpenseResponse
	err := ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[expenses.GroupExpense]{}
		spec.Model.ID = id
		spec.Model.CreatorProfileID = profileID
		spec.ForUpdate = true
		spec.PreloadRelations = []string{
			"Items",
			"OtherFees",
			"Payer",
			"Creator",
			"Items.Participants",
			"Items.Participants.Profile",
			"OtherFees.Participants",
			"OtherFees.Participants.Profile",
			"Participants",
			"Participants.Profile",
		}

		groupExpense, err := ges.getGroupExpense(ctx, spec)
		if err != nil {
			return err
		}

		if err = validateForConfirmation(groupExpense); err != nil {
			return err
		}

		updatedParticipants, err := ges.calculateUpdatedExpenseParticipants(ctx, groupExpense)
		if err != nil {
			return err
		}

		if !dryRun {
			groupExpense.Status = expenses.ConfirmedExpense
			groupExpense, err = ges.expenseRepo.Update(ctx, groupExpense)
			if err != nil {
				return err
			}
		}

		if err = ges.expenseRepo.SyncParticipants(ctx, groupExpense.ID, updatedParticipants); err != nil {
			return err
		}

		groupExpense.Participants = updatedParticipants

		response = mapper.GroupExpenseToResponse(groupExpense, profileID, "")

		return nil
	})
	return response, err
}

func validateForConfirmation(groupExpense expenses.GroupExpense) error {
	if groupExpense.Status == expenses.ConfirmedExpense {
		return ungerr.UnprocessableEntityError("already confirmed")
	}
	if len(groupExpense.Items) < 1 {
		return ungerr.UnprocessableEntityError("cannot confirm empty items")
	}
	if groupExpense.Status != expenses.ReadyExpense {
		return ungerr.UnprocessableEntityError("expense is not ready to confirm")
	}
	if !groupExpense.PayerProfileID.Valid {
		return ungerr.UnprocessableEntityError("no payer is selected")
	}
	return nil
}

func (ges *groupExpenseServiceImpl) calculateUpdatedExpenseParticipants(ctx context.Context, groupExpense expenses.GroupExpense) ([]expenses.ExpenseParticipant, error) {
	participantsMap := make(map[uuid.UUID]*expenses.ExpenseParticipant, len(groupExpense.Participants))
	for _, participant := range groupExpense.Participants {
		participant.ShareAmount = decimal.Zero
		participantsMap[participant.ParticipantProfileID] = &participant
	}

	for _, item := range groupExpense.Items {
		if len(item.Participants) < 1 {
			return nil, ungerr.UnprocessableEntityError(fmt.Sprintf("item %s does not have participants", item.Name))
		}
		for _, participant := range item.Participants {
			amountToAdd := item.TotalAmount().Mul(participant.Share)
			expenseParticipant, ok := participantsMap[participant.ProfileID]
			if !ok {
				return nil, ungerr.Unknownf("profile ID: %s is not found in expense participants", participant.ProfileID.String())
			}
			expenseParticipant.ShareAmount = expenseParticipant.ShareAmount.Add(amountToAdd)
		}
	}

	updatedParticipants := make([]expenses.ExpenseParticipant, 0, len(participantsMap))
	for _, participant := range participantsMap {
		updatedParticipants = append(updatedParticipants, *participant)
	}

	groupExpense.Participants = updatedParticipants

	updatedOtherFees, err := ges.calculateOtherFeeSplits(ctx, groupExpense)
	if err != nil {
		return nil, err
	}

	for _, fee := range updatedOtherFees {
		for _, participant := range fee.Participants {
			expenseParticipant, ok := participantsMap[participant.ProfileID]
			if !ok {
				return nil, ungerr.Unknown("missing participant profile from other fee")
			}
			expenseParticipant.ShareAmount = expenseParticipant.ShareAmount.Add(participant.ShareAmount)
		}
	}

	finalParticipants := make([]expenses.ExpenseParticipant, 0, len(participantsMap))
	for _, participant := range participantsMap {
		finalParticipants = append(finalParticipants, *participant)
	}

	return finalParticipants, nil
}

func (ges *groupExpenseServiceImpl) calculateOtherFeeSplits(ctx context.Context, groupExpense expenses.GroupExpense) ([]expenses.OtherFee, error) {
	var splitErr error

	mapperFunc := func(fee expenses.OtherFee) expenses.OtherFee {
		feeCalculator, ok := ges.feeCalculatorRegistry[fee.CalculationMethod]
		if !ok {
			splitErr = ungerr.Unknownf("unsupported calculation method: %s", fee.CalculationMethod)
			return expenses.OtherFee{}
		}

		if err := feeCalculator.Validate(fee, groupExpense); err != nil {
			splitErr = err
			return expenses.OtherFee{}
		}

		fee.Participants = feeCalculator.Split(fee, groupExpense)

		if err := ges.otherFeeRepository.SyncParticipants(ctx, fee.ID, fee.Participants); err != nil {
			splitErr = err
			return expenses.OtherFee{}
		}

		return fee
	}

	splitFees := ezutil.MapSlice(groupExpense.OtherFees, mapperFunc)

	return splitFees, splitErr
}

func (ges *groupExpenseServiceImpl) Delete(ctx context.Context, userProfileID, id uuid.UUID) error {
	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		expense, err := ges.GetUnconfirmedGroupExpenseForUpdate(ctx, userProfileID, id)
		if err != nil {
			return err
		}
		return ges.expenseRepo.Delete(ctx, expense)
	})
}

func (ges *groupExpenseServiceImpl) SyncParticipants(ctx context.Context, req dto.ExpenseParticipantsRequest) error {
	participants, profileIDs, err := ges.validateAndGetParticipants(ctx, req)
	if err != nil {
		return err
	}

	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		expense, err := ges.GetUnconfirmedGroupExpenseForUpdate(ctx, req.UserProfileID, req.GroupExpenseID)
		if err != nil {
			return err
		}

		expense.PayerProfileID = uuid.NullUUID{
			UUID:  req.PayerProfileID,
			Valid: true,
		}

		if _, err = ges.expenseRepo.Update(ctx, expense); err != nil {
			return err
		}

		if err := ges.expenseRepo.SyncParticipants(ctx, expense.ID, participants); err != nil {
			return err
		}

		return ges.expenseRepo.DeleteItemParticipants(ctx, expense.ID, profileIDs)
	})
}

func (ges *groupExpenseServiceImpl) validateAndGetParticipants(ctx context.Context, req dto.ExpenseParticipantsRequest) ([]expenses.ExpenseParticipant, []uuid.UUID, error) {
	participantSet := mapset.NewSet[uuid.UUID]()
	for _, pid := range req.ParticipantProfileIDs {
		participantSet.Add(pid)
	}

	if participantSet.Cardinality() != len(req.ParticipantProfileIDs) {
		return nil, nil, ungerr.UnprocessableEntityError("duplicate participant profile IDs given")
	}
	if !participantSet.Contains(req.PayerProfileID) {
		return nil, nil, ungerr.UnprocessableEntityError("payer profile ID must be one of the participant profile IDs")
	}

	for _, participantProfileID := range req.ParticipantProfileIDs {
		if participantProfileID == req.UserProfileID {
			continue
		}
		isFriends, _, err := ges.friendshipService.IsFriends(ctx, req.UserProfileID, participantProfileID)
		if err != nil {
			return nil, nil, err
		}
		if !isFriends {
			return nil, nil, ungerr.UnprocessableEntityError(appconstant.ErrNotFriends)
		}
	}

	if !participantSet.Contains(req.UserProfileID) {
		participantSet.Add(req.UserProfileID)
	}

	participantSlice := participantSet.ToSlice()

	return ezutil.MapSlice(participantSlice, func(id uuid.UUID) expenses.ExpenseParticipant {
		return expenses.ExpenseParticipant{ParticipantProfileID: id}
	}), participantSlice, nil
}

func (ges *groupExpenseServiceImpl) getGroupExpense(ctx context.Context, spec crud.Specification[expenses.GroupExpense]) (expenses.GroupExpense, error) {
	groupExpense, err := ges.expenseRepo.FindFirst(ctx, spec)
	if err != nil {
		return expenses.GroupExpense{}, err
	}
	if groupExpense.IsZero() {
		return expenses.GroupExpense{}, ungerr.NotFoundError(fmt.Sprintf("group expense with ID %s is not found", spec.Model.ID))
	}
	return groupExpense, nil
}

func (ges *groupExpenseServiceImpl) GetUnconfirmedGroupExpenseForUpdate(ctx context.Context, profileID, id uuid.UUID) (expenses.GroupExpense, error) {
	spec := crud.Specification[expenses.GroupExpense]{}
	spec.Model.ID = id
	if profileID != uuid.Nil {
		spec.Model.CreatorProfileID = profileID
	}
	spec.ForUpdate = true
	spec.PreloadRelations = []string{"Items", "Items.Participants"}
	groupExpense, err := ges.getGroupExpense(ctx, spec)
	if err != nil {
		return expenses.GroupExpense{}, err
	}
	if groupExpense.Status == expenses.ConfirmedExpense {
		return expenses.GroupExpense{}, ungerr.UnprocessableEntityError("expense already confirmed")
	}

	return groupExpense, nil
}

func (ges *groupExpenseServiceImpl) ParseFromBillText(ctx context.Context, msg message.ExpenseBillTextExtracted) error {
	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		expenseBill, err := ges.getPendingForProcessingExpenseBill(ctx, msg.ID)
		if err != nil {
			return err
		}
		if err := ges.parseFlow(ctx, expenseBill); err != nil {
			// Log error but do not return error (commit the transaction)
			logger.Errorf("error processing bill parsing: %v", err)
		}
		return nil
	})
}

func (ges *groupExpenseServiceImpl) parseFlow(ctx context.Context, expenseBill expenses.ExpenseBill) error {
	status, err := ges.processAndGetStatus(ctx, expenseBill)
	expenseBill.Status = status
	_, statusErr := ges.billRepo.Update(ctx, expenseBill)
	if statusErr != nil {
		return errors.Join(statusErr, err)
	}
	return err
}

func (ges *groupExpenseServiceImpl) processAndGetStatus(ctx context.Context, expenseBill expenses.ExpenseBill) (expenses.BillStatus, error) {
	expense, err := ges.GetUnconfirmedGroupExpenseForUpdate(ctx, uuid.Nil, expenseBill.GroupExpenseID)
	if err != nil {
		return expenses.FailedParsingBill, err
	}

	request, err := ges.parseExpenseBillTextToExpenseRequest(ctx, expenseBill.ExtractedText)
	if err != nil {
		if errors.Is(err, expenses.ErrExpenseNotDetected) {
			return expenses.NotDetectedBill, nil
		}
		return expenses.FailedParsingBill, err
	}

	if err = ges.UpdateDraft(ctx, expense, request); err != nil {
		return expenses.FailedParsingBill, err
	}

	return expenses.ParsedBill, nil
}

func (ges *groupExpenseServiceImpl) UpdateDraft(ctx context.Context, expense expenses.GroupExpense, request dto.NewGroupExpenseRequest) error {
	if err := ges.validate(request); err != nil {
		return err
	}

	mappedExpense := mapper.GroupExpenseRequestToEntity(request)
	expense.TotalAmount = mappedExpense.TotalAmount
	expense.ItemsTotal = mappedExpense.ItemsTotal
	expense.FeesTotal = mappedExpense.FeesTotal
	expense.Items = mappedExpense.Items
	expense.OtherFees = mappedExpense.OtherFees

	if expense.Description == "" {
		expense.Description = mappedExpense.Description
	}

	_, err := ges.expenseRepo.Update(ctx, expense)
	return err
}

func (ges *groupExpenseServiceImpl) validate(request dto.NewGroupExpenseRequest) error {
	if request.TotalAmount.IsZero() {
		return ungerr.UnprocessableEntityError(appconstant.ErrAmountZero)
	}

	calculatedFeeTotal := decimal.Zero
	calculatedSubtotal := decimal.Zero
	for _, item := range request.Items {
		calculatedSubtotal = calculatedSubtotal.Add(item.Amount.Mul(decimal.NewFromInt(int64(item.Quantity))))
	}
	for _, fee := range request.OtherFees {
		calculatedFeeTotal = calculatedFeeTotal.Add(fee.Amount)
	}
	if calculatedFeeTotal.Add(calculatedSubtotal).Cmp(request.TotalAmount) != 0 {
		return ungerr.UnprocessableEntityError(appconstant.ErrAmountMismatched)
	}
	if calculatedSubtotal.Cmp(request.Subtotal) != 0 {
		return ungerr.UnprocessableEntityError(appconstant.ErrAmountMismatched)
	}

	return nil
}

func (ges *groupExpenseServiceImpl) parseExpenseBillTextToExpenseRequest(ctx context.Context, text string) (dto.NewGroupExpenseRequest, error) {
	promptResponse, err := ges.llmService.Prompt(ctx, ges.buildSystemPrompt(), ges.buildUserPrompt(text))
	if err != nil {
		return dto.NewGroupExpenseRequest{}, err
	}
	if promptResponse == string(expenses.NotDetectedBill) {
		logger.Info("group expense not detected")
		return dto.NewGroupExpenseRequest{}, expenses.ErrExpenseNotDetected
	}

	var request dto.NewGroupExpenseRequest
	if err = json.Unmarshal([]byte(promptResponse), &request); err != nil {
		return dto.NewGroupExpenseRequest{}, ungerr.Wrap(err, "error unmarshaling to JSON")
	}

	return request, nil
}

func (ges *groupExpenseServiceImpl) buildSystemPrompt() string {
	return fmt.Sprintf(`You are an expert at parsing expense documents and receipts. 
Extract the expense information and return ONLY a valid JSON object in the following schema:

{
  "totalAmount": number,
  "subtotal": number,
  "description": string,
  "items": [
    {
      "name": string,
      "amount": number,   // price per unit
      "quantity": number
    }
  ],
  "otherFees": [
    {
      "name": string,
      "amount": number,
      "calculationMethod": "EQUAL_SPLIT" | "ITEMIZED_SPLIT"
    }
  ]
}

INSTRUCTIONS:
1. Return ONLY the JSON object, no explanations, no comments.
2. The JSON must be compact (no spaces, no line breaks, no pretty formatting).
3. totalAmount = subtotal + sum of otherFees.
4. subtotal = sum of (item.amount * item.quantity).
5. If subtotal is not explicitly mentioned, calculate it.
6. If quantity is not specified, assume 1.
7. Item.amount is always price per unit, not total for all units.
8. For otherFees:
   - Use "ITEMIZED_SPLIT" for percentage-based fees like tax or service charge, 
     because they should be distributed proportionally to the items each person ordered.
   - Use "EQUAL_SPLIT" only for true flat fees that apply equally regardless of items 
     (e.g., table charge, fixed booking fee).
9. All numeric values must be numbers, not strings.
10. Decimal separator can be "." or ",". Normalize both to "." in the output.
   - Example: "10,5" → 10.5
   - Example: "10.50" → 10.5
11. Thousands separators may appear as "." or "," in the input. Always remove them before parsing.
   - Example: "10.000" → 10000
   - Example: "10,000" → 10000
12. The final output must contain plain numeric values, with no thousands separators, and "." as the decimal separator.
13. If no clear expense information exists, return string "%s"`, expenses.NotDetectedBill)
}

func (ges *groupExpenseServiceImpl) buildUserPrompt(text string) string {
	return fmt.Sprintf("TEXT TO PARSE:\n%s", text)
}

func (ges *groupExpenseServiceImpl) getPendingForProcessingExpenseBill(ctx context.Context, id uuid.UUID) (expenses.ExpenseBill, error) {
	spec := crud.Specification[expenses.ExpenseBill]{}
	spec.Model.ID = id
	spec.ForUpdate = true
	expenseBill, err := ges.billRepo.FindFirst(ctx, spec)
	if err != nil {
		return expenses.ExpenseBill{}, err
	}
	if expenseBill.Status == expenses.ParsedBill {
		return expenses.ExpenseBill{}, ungerr.Unknownf("expense bill ID: %s already parsed", expenseBill.GroupExpenseID)
	}
	return expenseBill, nil
}
