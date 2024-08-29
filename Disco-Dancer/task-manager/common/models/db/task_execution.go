package db

import (
	"encoding/json"
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"

	appmodels "geico.visualstudio.com/Billing/plutus/task-manager-common/models/app"
)

type TaskExecution struct {
	TaskExecutionId     int
	TaskId              int
	ExecutionDate       time.Time
	ExecutionParameters appmodels.TaskExecutionParemeters
	StartTime           time.Time
	EndTime             time.Time
	Status              enums.TaskStatus
	RecordsProcessed    int
	ErrorDetails        json.RawMessage
}
