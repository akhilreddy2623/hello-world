package repository

import (
	"context"
	"time"

	"geico.visualstudio.com/Billing/plutus/database"
	dbModels "geico.visualstudio.com/Billing/plutus/task-manager-common/models/db"
	"github.com/jackc/pgx/v5"
)

const tskMgrWrkUser = "task-manager-wrk"

type TaskScheduleRepositoryInterface interface {
	GetDueSchedules() ([]*dbModels.TaskSchedule, error)
}

type TaskScheduleRepository struct {
}

func (TaskScheduleRepository) GetDueSchedules() ([]*dbModels.TaskSchedule, error) {

	tasksSchedules, err := getTaskSchedulesFromDb()
	if err != nil {
		repositoryLog.Error(context.Background(), err, "getTaskSchedules from db failed")
		return nil, err
	}

	return tasksSchedules, nil
}

func (TaskScheduleRepository) InsertTaskExecutionRequestInDb(taskExecution *dbModels.TaskExecution) (pgx.Tx, error) {
	addTaskExecutionRequestQuery := `INSERT INTO public.task_execution
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

	// Open a transaction
	transaction, err := database.NewDbContext().Database.Begin(context.Background())

	if err != nil {
		return nil, err
	}

	// Execute the SQL statement
	_, err = transaction.Exec(
		context.Background(),
		addTaskExecutionRequestQuery,
		taskExecution.TaskId,
		taskExecution.ExecutionDate,
		taskExecution.ExecutionParameters,
		taskExecution.StartTime,
		taskExecution.Status,
		time.Now(),
		tskMgrWrkUser,
		time.Now(),
		tskMgrWrkUser,
	)

	if err != nil {
		transaction.Rollback(context.Background())
		repositoryLog.Error(context.Background(), err, "InsertTaskExecutionRequestInDb - error inserting record in task_execution table")
		return nil, err
	}

	return transaction, nil
}

func (TaskScheduleRepository) UpdateNextRunOnTaskSchedule(nextRunToUpdate time.Time, taskId int, transaction pgx.Tx) error {
	updateTaskScheduleQuery := `UPDATE public.task_schedule
	SET
		"NextRun" = $1,
		"UpdatedDate" = $2,
		"UpdatedBy" = $3
	WHERE
		"TaskId" = $4`

	// Execute the SQL statement
	_, err := transaction.Exec(
		context.Background(),
		updateTaskScheduleQuery,
		nextRunToUpdate,
		time.Now(),
		tskMgrWrkUser,
		taskId,
	)

	if err != nil {
		transaction.Rollback(context.Background())
		repositoryLog.Error(context.Background(), err, "UpdateTaskSchedule - error updating task schedule for next run")
		return err
	}

	return nil

}

func getTaskSchedulesFromDb() ([]*dbModels.TaskSchedule, error) {
	var taskSchedules []*dbModels.TaskSchedule

	getTaskSchedulesQuery :=
		`SELECT 
			"ScheduleId", 
			"TaskId", 
			"ExecutionParameters", 
			"Increment", 
			"MaxRunTime", 
			"Schedule", 
			COALESCE("NextRun", timestamp '2000-01-01 00:00:00'), 
			"Status"	
		FROM 
			public.task_schedule
		Where
			"NextRun" < NOW()
		Order By 1 ASC`

	rows, err := database.NewDbContext().Database.Query(
		context.Background(),
		getTaskSchedulesQuery)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "getTaskSchedulesFromDb - error in task schedules from task_schedule")
		return nil, err
	}

	for rows.Next() {
		var taskSchedule dbModels.TaskSchedule
		err := rows.Scan(
			&taskSchedule.ScheduleId,
			&taskSchedule.TaskId,
			&taskSchedule.ExecutionParameters,
			&taskSchedule.Increment,
			&taskSchedule.MaxRunTime,
			&taskSchedule.Schedule,
			&taskSchedule.NextRun,
			&taskSchedule.Status)

		if err != nil {
			repositoryLog.Error(context.Background(), err, "getTasksSchdulesFromDb - error in scanning task schedules")
			return nil, err
		}
		taskSchedules = append(taskSchedules, &taskSchedule)
	}

	return taskSchedules, nil
}
