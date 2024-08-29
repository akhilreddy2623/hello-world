package db

import (
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
	appModels "geico.visualstudio.com/Billing/plutus/task-manager-common/models/app"
)

type TaskSchedule struct {
	ScheduleId          int
	TaskId              int
	ExecutionParameters TaskScheduleExecutionParameters
	Increment           int
	MaxRunTime          int
	Schedule            string
	NextRun             time.Time
	Status              enums.TaskStatus
}

type TaskScheduleExecutionParameters struct {
	Day                 string
	Year                string
	Month               string
	TaskId              int
	ExecutionParameters appModels.TaskExecutionParemeters
}
