package enums

import (
	"strings"
)

type TaskStatus int

const (
	NoneTaskStatus TaskStatus = iota
	TaskStaged
	TaskInprogress
	TaskCompleted
	TaskErrored
	TaskOnHold
)

func (t TaskStatus) String() string {
	return [...]string{
		"",
		"staged",
		"inprogress",
		"completed",
		"errored",
		"onhold"}[t]
}

func (t TaskStatus) EnumIndex() int {
	return int(t)
}

func GetTaskStatusEnum(taskstatus string) TaskStatus {
	switch strings.ToLower(taskstatus) {
	case "staged":
		return TaskStaged
	case "inprogress":
		return TaskInprogress
	case "completed":
		return TaskCompleted
	case "errored":
		return TaskErrored
	case "ohhold":
		return TaskOnHold
	}
	return NoneTaskStatus
}

func GetTaskStatusEnumByIndex(taskStatus int) TaskStatus {
	switch taskStatus {
	case 1:
		return TaskStaged
	case 2:
		return TaskInprogress
	case 3:
		return TaskCompleted
	case 4:
		return TaskErrored
	case 5:
		return TaskOnHold
	}
	return NoneTaskStatus
}
