package messaging

import (
	"strings"
	"time"
)

const TimeFormat = "2006-01-02"

type ExecuteTaskRequest struct {
	Version             int                 `json:"version"`
	Component           string              `json:"component"`
	TaskName            string              `json:"taskName"`
	TaskDate            JsonDate            `json:"taskDate"`
	TaskExecutionId     int                 `json:"taskExecutionId"`
	ExecutionParameters ExecutionParameters `json:"executionParameters"`
}

type ExecutionParameters struct {
	PaymentMethodType  string `json:"paymentMethodType"`
	PaymentRequestType string `json:"paymentRequestType"`
	PaymentFrequency   string `json:"paymentFrequency"`
	WorkdayFeed        string `json:"workdayFeed"`
}

type JsonDate time.Time

func (j *JsonDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse(TimeFormat, s)
	if err != nil {
		return err
	}
	*j = JsonDate(t)
	return nil
}

func (j JsonDate) MarshalJSON() ([]byte, error) {
	t := time.Time(j)
	formatted := t.Format(TimeFormat)
	jsonStr := "\"" + formatted + "\""
	return []byte(jsonStr), nil
}
