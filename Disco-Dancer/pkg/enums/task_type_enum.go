package enums

import (
	"strings"
)

type TaskType int

const (
	NoneTaskType TaskType = iota
	ProcessPayments
	SettleACHPayments
	SettleCardPayments
	SendWorkdayData
)

func (t TaskType) String() string {
	return [...]string{
		"",
		"processpayments",
		"settleachpayments",
		"settlecardpayments",
		"sendworkdaydata"}[t]
}

func (t TaskType) EnumIndex() int {
	return int(t)
}

func GetTaskTypeEnum(taskType string) TaskType {
	switch strings.ToLower(taskType) {
	case "processpayments":
		return ProcessPayments
	case "settleachpayments":
		return SettleACHPayments
	case "settlecardpayments":
		return SettleCardPayments
	case "sendworkdaydata":
		return SendWorkdayData
	}
	return NoneTaskType
}
