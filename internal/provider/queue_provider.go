package provider

import (
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/meq"
)

type Queues struct {
	ExpenseBillUploaded      meq.TaskQueue[message.ExpenseBillUploaded]
	ExpenseBillTextExtracted meq.TaskQueue[message.ExpenseBillTextExtracted]
}

func ProvideQueues(db meq.DB) *Queues {
	return &Queues{
		ExpenseBillUploaded:      meq.NewTaskQueue[message.ExpenseBillUploaded](logger.Global, db),
		ExpenseBillTextExtracted: meq.NewTaskQueue[message.ExpenseBillTextExtracted](logger.Global, db),
	}
}
