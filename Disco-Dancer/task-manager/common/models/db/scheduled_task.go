package db

type ScheduledTask struct {
	TaskId    int
	TaskName  string
	Component string
	DependsOn []int
	IsActive  bool
}
