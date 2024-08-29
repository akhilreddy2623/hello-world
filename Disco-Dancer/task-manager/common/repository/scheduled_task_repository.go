package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"github.com/jackc/pgx/v5"

	dbmodels "geico.visualstudio.com/Billing/plutus/task-manager-common/models/db"
)

const user = "task-manager" //TODO: should this be system name or user calling the API?, Will we be get this detail from APISIX Gateway?

var repositoryLog = logging.GetLogger("task-manager-repository")

// TODO: Learn how to generate the mock file

//go:generate mockery --name ScheduledTaskRepositoryInterface
type ScheduledTaskRepositoryInterface interface {
	AreTaskDependenciesMet(taskId int, executionDate time.Time) (*dbmodels.ScheduledTask, error)
	PersistTaskExecutionRequestInDB(taskExecution *dbmodels.TaskExecution) error
	UpdateTaskStatus(executeTaskResponses commonMessagingModels.ExecuteTaskResponse) error
}

type ScheduledTaskRepository struct {
}

func (ScheduledTaskRepository) AreTaskDependenciesMet(taskId int, executionDate time.Time) (*dbmodels.ScheduledTask, error) {
	scheduledTask, err := getScheduledTask(taskId)
	if err != nil {
		return nil, err
	}
	for _, dependency := range scheduledTask.DependsOn {
		isComplete, err := isTaskComplete(dependency, executionDate)
		if err != nil {
			return nil, err
		}
		if !isComplete {
			return nil, errors.New("TaskId " + strconv.Itoa(dependency) + " is not complete")
		}
	}
	return &scheduledTask, nil
}

func (ScheduledTaskRepository) PersistTaskExecutionRequestInDB(taskExecution *dbmodels.TaskExecution) error {

	addTaskExecutionRequest := `INSERT INTO public.task_execution
	("TaskId",
	 "ExecutionDate",
	 "ExecutionParameters", 
	 "StartTime", 
	 "Status", 
	 "CreatedDate",
	 "CreatedBy", 
	 "UpdatedDate",
	 "UpdatedBy"
	)
	VALUES(
	$1, $2, $3,	$4, $5, $6, $7,	$8, $9
	)
	RETURNING "TaskExecutionId"`

	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		addTaskExecutionRequest,
		taskExecution.TaskId,
		taskExecution.ExecutionDate,
		taskExecution.ExecutionParameters,
		taskExecution.StartTime,
		taskExecution.Status,
		time.Now(),
		user,
		time.Now(),
		user,
	)

	err := row.Scan(&taskExecution.TaskExecutionId)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "error inserting record in task_execution table")
		return err
	}
	return nil
}

func getScheduledTask(taskId int) (dbmodels.ScheduledTask, error) {
	var scheduledTask dbmodels.ScheduledTask

	getDependenciesQuery := `select "TaskId", "TaskName", "Component", "DependsOn", "IsActive" from public.scheduled_task where "TaskId"=$1`
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getDependenciesQuery,
		taskId)
	err := row.Scan(&scheduledTask.TaskId, &scheduledTask.TaskName, &scheduledTask.Component, &scheduledTask.DependsOn, &scheduledTask.IsActive)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in getting scheduled_task from database")
	}

	return scheduledTask, err
}

func isTaskComplete(taskId int, executionDate time.Time) (bool, error) {
	var status int

	getStatusQuery := `select "Status" from public.task_execution where "TaskId"=$1 and "ExecutionDate" = $2`
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getStatusQuery,
		taskId,
		executionDate)
	err := row.Scan(&status)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in getting task status from database")
	}

	return enums.GetTaskStatusEnumByIndex(status) == enums.TaskCompleted, err
}

func (ScheduledTaskRepository) UpdateTaskStatus(executeTaskResponses commonMessagingModels.ExecuteTaskResponse) error {

	transaction, err := database.NewDbContext().Database.Begin(context.Background())

	defer transaction.Rollback(context.Background())
	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in starting transaction for UpdateTaskStatus TaskExecutionId : '%d'", executeTaskResponses.TaskExecutionId)
		return err
	}

	err = updateTaskInfoStatus(transaction, executeTaskResponses)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in updating task status in database")
		return err
	}
	transaction.Commit(context.Background())
	return nil
}

func updateTaskInfoStatus(transaction pgx.Tx, executeTaskResponses commonMessagingModels.ExecuteTaskResponse) error {
	updateExecutionQuery := `UPDATE public.task_execution SET "Status" = $1,"EndTime"=$2, "RecordsProcessed"=$3,"ErrorDetails"=$4, "UpdatedBy"=$5,"UpdatedDate"=$6 WHERE "TaskExecutionId" = $7`

	var enumStatus int

	if enums.GetTaskStatusEnum(executeTaskResponses.Status) == enums.TaskStatus(enums.TaskErrored) {
		enumStatus = int(enums.TaskErrored)
	} else if enums.GetTaskStatusEnum(executeTaskResponses.Status) == enums.TaskStatus(enums.TaskInprogress) {
		enumStatus = int(enums.TaskInprogress)
	} else if enums.GetTaskStatusEnum(executeTaskResponses.Status) == enums.TaskStatus(enums.TaskCompleted) {
		enumStatus = int(enums.TaskCompleted)
	}

	_, err := transaction.Exec(
		context.Background(),
		updateExecutionQuery,
		enumStatus,
		time.Now(),
		executeTaskResponses.ProcessedRecordsCount,
		executeTaskResponses.ErrorDetails,
		user,
		time.Now(),
		executeTaskResponses.TaskExecutionId)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error updating task_execution table")
		return err
	}

	return nil
}
