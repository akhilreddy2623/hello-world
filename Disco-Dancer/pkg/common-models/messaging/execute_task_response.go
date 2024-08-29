package messaging

import commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"

type ExecuteTaskResponse struct {
	Version               int                            `json:"version"`
	TaskExecutionId       int                            `json:"taskExecutionId"`
	Status                string                         `json:"status"`
	ProcessedRecordsCount int                            `json:"processedRecordsCount"`
	ErrorDetails          *commonAppModels.ErrorResponse `json:"errorDetails,omitempty"`
}
