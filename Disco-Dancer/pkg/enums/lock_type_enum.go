package enums

import "strings"

// LockTypeEnum is an enum type for the type of lock

type LockType int

const (
	NoneLockType LockType = iota
	ConsolidatePaymentRequestsLock
	CreateACHOneTimeFileLock
)

func (t LockType) String() string {
	return [...]string{
		"",
		"consolidatepaymentrequestslock",
		"createachonetimefilelock"}[t]
}

func (t LockType) EnumIndex() int {
	return int(t)
}

func GetLockTypeEnum(taskType string) LockType {
	switch strings.ToLower(taskType) {
	case "consolidatepaymentrequestslock":
		return ConsolidatePaymentRequestsLock
	case "createachonetimefilelock":
		return CreateACHOneTimeFileLock
	}
	return NoneLockType
}
