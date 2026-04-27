package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/core/service/langfuse"
	"github.com/itsLeonB/cashback/internal/core/service/llm"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/cashback/internal/domain/service/expense"
	"github.com/itsLeonB/cashback/internal/domain/service/expense/billparse"
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
	calculationSvc        expense.CalculationService
	imageSvc              storage.ImageService
	taskQueue             queue.TaskQueue
	langfuseClient        langfuse.Client
}

func NewGroupExpenseService(
	friendshipService FriendshipService,
	expenseRepo repository.GroupExpenseRepository,
	transactor crud.Transactor,
	feeCalculatorRegistry map[expenses.FeeCalculationMethod]fee.FeeCalculator,
	otherFeeRepository repository.OtherFeeRepository,
	billRepo crud.Repository[expenses.ExpenseBill],
	llmService llm.LLMService,
	imageSvc storage.ImageService,
	taskQueue queue.TaskQueue,
	langfuseClient langfuse.Client,
) GroupExpenseService {
	return &groupExpenseServiceImpl{
		friendshipService,
		expenseRepo,
		transactor,
		feeCalculatorRegistry,
		otherFeeRepository,
		billRepo,
		llmService,
		expense.NewCalculationService(),
		imageSvc,
		taskQueue,
		langfuseClient,
	}
}

func (ges *groupExpenseServiceImpl) CreateDraft(ctx context.Context, userProfileID uuid.UUID, description string) (dto.GroupExpenseResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.CreateDraft")
	defer span.End()

	newDraftExpense := expenses.GroupExpense{
		CreatorProfileID: userProfileID,
		Description:      description,
		Status:           expenses.DraftExpense,
	}

	insertedDraftExpense, err := ges.expenseRepo.Insert(ctx, newDraftExpense)
	if err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	return mapper.GroupExpenseToResponse(insertedDraftExpense, userProfileID, "", false), nil
}

func (ges *groupExpenseServiceImpl) GetAll(ctx context.Context, userProfileID uuid.UUID, ownership expenses.ExpenseOwnership, status expenses.ExpenseStatus) ([]dto.GroupExpenseResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.GetAll")
	defer span.End()

	groupExpenses, err := ges.expenseRepo.FindAllByOwnership(ctx, userProfileID, ownership, status, -1)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(groupExpenses, mapper.GroupExpenseSimpleMapper(userProfileID, "", false)), nil
}

func (ges *groupExpenseServiceImpl) GetDetails(ctx context.Context, id, userProfileID uuid.UUID) (dto.GroupExpenseResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.GetDetails")
	defer span.End()

	spec := crud.Specification[expenses.GroupExpense]{}
	spec.Model.ID = id
	spec.PreloadRelations = spec.Model.ForDisplayRelations()

	groupExpense, err := ges.getGroupExpense(ctx, spec)
	if err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	// Check if user has permission to view this expense (creator or participant)
	isCreator := groupExpense.CreatorProfileID == userProfileID
	isParticipant := false
	for _, participant := range groupExpense.Participants {
		if participant.ParticipantProfileID == userProfileID {
			isParticipant = true
			break
		}
	}

	if !isCreator && !isParticipant {
		return dto.GroupExpenseResponse{}, ungerr.NotFoundError("expense not found")
	}

	var billURL string
	if !groupExpense.Bill.IsZero() {
		billURL, err = ges.imageSvc.GetURL(ObjectKeyToFileID(groupExpense.Bill.ImageName))
		if err != nil {
			logger.Errorf("error retrieving bill image URL: %v", err)
		}
	}

	return mapper.GroupExpenseToResponse(groupExpense, userProfileID, billURL, groupExpense.Status == expenses.ConfirmedExpense), nil
}

func (ges *groupExpenseServiceImpl) ConfirmDraft(ctx context.Context, id, profileID uuid.UUID, dryRun bool) (dto.ExpenseConfirmationResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.ConfirmDraft")
	defer span.End()

	var response dto.ExpenseConfirmationResponse
	err := ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[expenses.GroupExpense]{}
		spec.Model.ID = id
		spec.Model.CreatorProfileID = profileID
		spec.ForUpdate = true
		spec.PreloadRelations = spec.Model.ForCalculationRelations()

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

		if !dryRun {
			if err = ges.taskQueue.Enqueue(ctx, message.ExpenseConfirmed{ID: id}); err != nil {
				return err
			}
		}

		response = mapper.ToConfirmationResponse(groupExpense, profileID)
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
			amountToAdd := participant.AllocatedAmount
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
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.Delete")
	defer span.End()

	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		expense, err := ges.GetUnconfirmedForUpdate(ctx, userProfileID, id)
		if err != nil {
			return err
		}
		return ges.expenseRepo.Delete(ctx, expense)
	})
}

func (ges *groupExpenseServiceImpl) SyncParticipants(ctx context.Context, req dto.ExpenseParticipantsRequest) error {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.SyncParticipants")
	defer span.End()

	participants, profileIDs, err := ges.validateAndGetParticipants(ctx, req)
	if err != nil {
		return err
	}

	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		expense, err := ges.GetUnconfirmedForUpdate(ctx, req.UserProfileID, req.GroupExpenseID)
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
	// --- Dedup check ---
	participantSet := mapset.NewSet(req.ParticipantProfileIDs...)
	if participantSet.Cardinality() != len(req.ParticipantProfileIDs) {
		return nil, nil, ungerr.UnprocessableEntityError("duplicate participant profile IDs given")
	}
	if !participantSet.Contains(req.PayerProfileID) {
		return nil, nil, ungerr.UnprocessableEntityError("payer profile ID must be one of the participant profile IDs")
	}

	if err := ges.checkFriendships(ctx, req.ParticipantProfileIDs, req.UserProfileID); err != nil {
		return nil, nil, err
	}

	if req.ProxyByProfileIDs != nil {
		if err := ges.validateProxies(participantSet, req.ProxyByProfileIDs, req.PayerProfileID); err != nil {
			return nil, nil, err
		}
	}

	// --- Build participant slice ---
	participantSlice := req.ParticipantProfileIDs
	if !participantSet.Contains(req.UserProfileID) {
		participantSlice = append(participantSlice, req.UserProfileID)
	}

	participants := ezutil.MapSlice(participantSlice, func(id uuid.UUID) expenses.ExpenseParticipant {
		proxyID, ok := uuid.Nil, false
		if req.ProxyByProfileIDs != nil {
			proxyID, ok = req.ProxyByProfileIDs[id]
		}
		return expenses.ExpenseParticipant{
			ParticipantProfileID: id,
			ProxyProfileID:       uuid.NullUUID{UUID: proxyID, Valid: ok},
		}
	})

	return participants, participantSlice, nil
}

func (ges *groupExpenseServiceImpl) validateProxies(participantSet mapset.Set[uuid.UUID], proxyByProfileIDs map[uuid.UUID]uuid.UUID, payerProfileID uuid.UUID) error {
	for id, proxyID := range proxyByProfileIDs {
		if proxyID == id {
			return ungerr.UnprocessableEntityError("proxy cannot be the same as participant")
		}
		if proxyID == payerProfileID {
			return ungerr.UnprocessableEntityError("proxy cannot be the payer")
		}
		if id == payerProfileID {
			return ungerr.UnprocessableEntityError("payer cannot have a proxy")
		}
		if _, chained := proxyByProfileIDs[proxyID]; chained {
			return ungerr.UnprocessableEntityError(fmt.Sprintf("proxy %s cannot itself have a proxy (chaining not allowed)", proxyID))
		}
		if !participantSet.Contains(proxyID) {
			return ungerr.UnprocessableEntityError(fmt.Sprintf("proxy for participant %s does not exist in participants list", id))
		}
	}
	return nil
}

func (ges *groupExpenseServiceImpl) checkFriendships(ctx context.Context, participantProfileIDs []uuid.UUID, userProfileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.checkFriendships")
	defer span.End()

	for _, pid := range participantProfileIDs {
		if pid == userProfileID {
			continue
		}
		isFriends, _, err := ges.friendshipService.IsFriends(ctx, userProfileID, pid)
		if err != nil {
			return err
		}
		if !isFriends {
			return ungerr.UnprocessableEntityError(appconstant.ErrNotFriends)
		}
	}
	return nil
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

func (ges *groupExpenseServiceImpl) GetUnconfirmedForUpdate(ctx context.Context, profileID, id uuid.UUID) (expenses.GroupExpense, error) {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.GetUnconfirmed")
	defer span.End()

	spec := crud.Specification[expenses.GroupExpense]{}
	spec.Model.ID = id
	if profileID != uuid.Nil {
		spec.Model.CreatorProfileID = profileID
	}
	spec.ForUpdate = true
	spec.PreloadRelations = []string{"Items", "Items.Participants", "Bill"}
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
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.ParseFromBillText")
	defer span.End()

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
	request, err := ges.parseExpenseBillTextToExpenseRequest(ctx, expenseBill.ExtractedText)
	if err != nil {
		if errors.Is(err, expenses.ErrExpenseNotDetected) {
			return expenses.NotDetectedBill, nil
		}
		return expenses.FailedParsingBill, err
	}

	request.Items = slices.DeleteFunc(request.Items, func(item dto.NewExpenseItemRequest) bool { return item.Amount.Equal(decimal.Zero) })
	request.OtherFees = slices.DeleteFunc(request.OtherFees, func(fee dto.NewOtherFeeRequest) bool { return fee.Amount.Equal(decimal.Zero) })

	expense, err := ges.GetUnconfirmedForUpdate(ctx, uuid.Nil, expenseBill.GroupExpenseID)
	if err != nil {
		return expenses.FailedParsingBill, err
	}

	if err = ges.UpdateDraft(ctx, expense, request); err != nil {
		return expenses.FailedParsingBill, err
	}

	return expenses.ParsedBill, nil
}

func (ges *groupExpenseServiceImpl) UpdateDraft(ctx context.Context, expense expenses.GroupExpense, request dto.NewGroupExpenseRequest) error {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.UpdateDraft")
	defer span.End()

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

func (ges *groupExpenseServiceImpl) Recalculate(ctx context.Context, userProfileID, groupExpenseID uuid.UUID, amountChanged bool) error {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.Recalculate")
	defer span.End()

	groupExpense, err := ges.GetUnconfirmedForUpdate(ctx, userProfileID, groupExpenseID)
	if err != nil {
		return err
	}

	recalculatedExpense, recalculated, err := ges.calculationSvc.RecalculateExpense(groupExpense, amountChanged)
	if err != nil {
		return err
	}
	if !recalculated {
		return nil
	}

	_, err = ges.expenseRepo.Update(ctx, recalculatedExpense)
	return err
}

func (ges *groupExpenseServiceImpl) GetRecent(ctx context.Context, profileID uuid.UUID) ([]dto.GroupExpenseResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.GetRecent")
	defer span.End()

	// Get recent expenses (both owned and participating) with DB-level limit
	expenses, err := ges.expenseRepo.FindRecentByProfileID(ctx, profileID, 5)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(expenses, mapper.GroupExpenseSimpleMapper(profileID, "", false)), nil
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

func (ges *groupExpenseServiceImpl) parseExpenseBillTextToExpenseRequest(
	ctx context.Context, text string,
) (dto.NewGroupExpenseRequest, error) {
	// 1. Pre-process and normalize numeric strings
	normalizedText := billparse.Normalize(text)

	p := billparse.ActiveBillParsePrompt

	prompt, err := ges.langfuseClient.GetPrompt(ctx, p.PromptName, p.GetOptions())
	if err != nil {
		return dto.NewGroupExpenseRequest{}, err
	}

	msgs, err := prompt.Compile(p.CompileVars(string(expenses.NotDetectedBill), normalizedText))
	if err != nil {
		return dto.NewGroupExpenseRequest{}, err
	}

	llmMsgs := make([]llm.ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		llmMsgs = append(llmMsgs, llm.ChatMessage{Role: m.Role, Content: m.Content})
	}

	raw, err := ges.llmService.Chat(ctx, llmMsgs)
	if err != nil {
		return dto.NewGroupExpenseRequest{}, err
	}
	logger.Debugf("prompt response: %s", raw)

	if strings.TrimSpace(raw) == string(expenses.NotDetectedBill) {
		logger.Info("group expense not detected")
		return dto.NewGroupExpenseRequest{}, expenses.ErrExpenseNotDetected
	}

	return p.ParseResponse(raw)
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

func (ges *groupExpenseServiceImpl) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (expenses.GroupExpense, error) {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.GetByID")
	defer span.End()

	spec := crud.Specification[expenses.GroupExpense]{}
	spec.Model.ID = id
	spec.ForUpdate = forUpdate
	spec.PreloadRelations = spec.Model.ForCalculationRelations()

	return ges.getGroupExpense(ctx, spec)
}

func (ges *groupExpenseServiceImpl) ConstructNotifications(ctx context.Context, msg message.ExpenseConfirmed) ([]entity.Notification, error) {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.ConstructNotifications")
	defer span.End()

	expense, err := ges.GetByID(ctx, msg.ID, false)
	if err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(message.ExpenseConfirmedMetadata{CreatorName: expense.Creator.Name})
	if err != nil {
		return nil, ungerr.Wrap(err, "error marshaling metadata to json")
	}

	notifications := make([]entity.Notification, 0, len(expense.Participants))
	for _, participant := range expense.Participants {
		if participant.ParticipantProfileID != expense.CreatorProfileID {
			notifications = append(notifications, entity.Notification{
				ProfileID:  participant.ParticipantProfileID,
				Type:       msg.Type(),
				EntityType: "group-expense",
				EntityID:   msg.ID,
				Metadata:   metadata,
			})
		}
	}

	return notifications, nil
}

func (ges *groupExpenseServiceImpl) ProcessCallback(ctx context.Context, id uuid.UUID, callbackFn func(context.Context, expenses.GroupExpense) error) error {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseService.ProcessCallback")
	defer span.End()

	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		expense, err := ges.GetByID(ctx, id, true)
		if err != nil {
			return err
		}

		// Idempotency guard
		if expense.Processed {
			return nil
		}

		if expense.Status != expenses.ConfirmedExpense {
			return ungerr.Unknown("group expense is not confirmed")
		}
		if len(expense.Participants) < 1 {
			return ungerr.Unknown("no participants to process")
		}

		if err = callbackFn(ctx, expense); err != nil {
			return err
		}

		expense.Processed = true
		_, err = ges.expenseRepo.Update(ctx, expense)
		return err
	})
}
