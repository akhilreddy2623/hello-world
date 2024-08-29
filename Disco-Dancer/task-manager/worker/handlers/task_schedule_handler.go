package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	dbModels "geico.visualstudio.com/Billing/plutus/task-manager-common/models/db"

	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/task-manager-common/repository"
	"github.com/jackc/pgx/v5"
	"github.com/robfig/cron"
)

var ExecuteTaskRequestsTopic string
var TaskLookupInterval int

func InitTaskSchedules() {
	configHandler := commonFunctions.GetConfigHandler()
	ExecuteTaskRequestsTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	TaskLookupInterval = configHandler.GetInt("TaskSchedules.PollingIntervalInMinutes", 0)
}

func CheckAndRunTaskSchedules() {

	for {

		time.Sleep(time.Duration(TaskLookupInterval * int(time.Minute)))

		// Get the task schedules from the table task_schedule
		taskSchedulesFromDb, err := getTaskSchedulesFromDb()

		if err != nil {
			log.Error(context.Background(), err, "Error getting master task schedules from database")
		} else {
			log.Info(context.Background(), "master task schedules retrieved from database")
		}

		for _, eachTaskSchedule := range taskSchedulesFromDb {

			// Compare the cron schedule with a specific time
			// Get the next scheduled time in EST
			jobScheduleTime, currentTimeEST, nextRunFromDbInEST, err := getAllTimesInESTForComparison(eachTaskSchedule)

			if err != nil {
				log.Error(context.Background(), err, "Error getting time from cron in EST from method getTimeFromCronInEST")
			}

			log.Info(context.Background(), "jobScheduleTime: %v, currentTime: %v, nextRunFromDBInEST: %v", jobScheduleTime, currentTimeEST, nextRunFromDbInEST)

			// Compare the next scheduled time with the specific time in EST
			if jobScheduleTime.Before(nextRunFromDbInEST) && jobScheduleTime.Before(currentTimeEST) {
				log.Info(context.Background(), "Cron schedule match the job run time")
				// Perform the task after conditins are met
				err = executeTaskScheduledJob(eachTaskSchedule)
				if err != nil {
					log.Error(context.Background(), err, "Error performing scheduled task")
				}
			} else {
				log.Info(context.Background(), "Cron schedule does not match the job run time: %v", eachTaskSchedule.NextRun)
			}
		}
	}
}

func executeTaskScheduledJob(s *dbModels.TaskSchedule) error {

	// Check if task depedencis are met or not
	// If task dependicies are not met then increment the nextrun
	// Else publish the message to executeTaskRequest topic and increment the NextRun to Day + 1

	log.Info(context.Background(), "Performing scheduled task")
	scheduledTaskDbRecord, err := areTaskDependenciesMet(s)
	log.Info(context.Background(), "Dependency Check Done ")

	if err != nil {
		log.Error(context.Background(), err, "Error checking task dependencies for TaskId: %v", s.TaskId)
		return err
	}

	// Insert the record in task_execution table by opening a transaction
	taskExecutionDbRecord, transaction, err := insertTaskExecutionRequestRecordInDb(s)
	if err != nil {
		log.Error(context.Background(), err, "Error inserting task execution request in db")
	}

	// If insert in above statement is successful then use the same transaction to update the schedule to next day
	nextRunToUpdate := s.NextRun.AddDate(0, 0, 1)
	err = updateNextRunOnTaskSchedule(nextRunToUpdate, s.TaskId, transaction)
	if err != nil {
		log.Error(context.Background(), err, "Error updating task schedule")
	} else {
		// Publish the message to executeTaskRequest topic
		messagingError := publishMessageToTopic(taskExecutionDbRecord, *scheduledTaskDbRecord)
		if messagingError != nil {
			log.Error(context.Background(), messagingError, "Error publishing message to topic executeTaskRequest for task id %v", s.TaskId)
		} else {
			transaction.Commit(context.Background())
		}
	}

	return nil
}

func updateNextRunOnTaskSchedule(nextRunToUpdate time.Time, taskId int, transaction pgx.Tx) error {
	var scheduledTaskRepository repository.TaskScheduleRepository = repository.TaskScheduleRepository{}
	err := scheduledTaskRepository.UpdateNextRunOnTaskSchedule(nextRunToUpdate, taskId, transaction)
	return err
}

func publishMessageToTopic(taskExecutionDbRecord *dbModels.TaskExecution, scheduledTask dbModels.ScheduledTask) error {

	executeTaskRequest := commonMessagingModels.ExecuteTaskRequest{
		Version:         1,
		Component:       scheduledTask.Component,
		TaskName:        scheduledTask.TaskName,
		TaskDate:        commonMessagingModels.JsonDate(taskExecutionDbRecord.ExecutionDate),
		TaskExecutionId: taskExecutionDbRecord.TaskExecutionId,
		ExecutionParameters: commonMessagingModels.ExecutionParameters{
			PaymentMethodType:  taskExecutionDbRecord.ExecutionParameters.PaymentMethodType.String(),
			PaymentRequestType: taskExecutionDbRecord.ExecutionParameters.PaymentRequestType.String(),
			PaymentFrequency:   taskExecutionDbRecord.ExecutionParameters.PaymentFrequency.String(),
		},
	}

	executeTaskRequestJson, marshalError := json.Marshal(executeTaskRequest)
	if marshalError != nil {
		log.Error(context.Background(), marshalError, "error in marshalling schedulePaymentRequest")
		return marshalError
	}
	publishError := messaging.KafkaPublish(ExecuteTaskRequestsTopic, string(executeTaskRequestJson))
	if publishError != nil {
		log.Error(context.Background(), publishError, "error in marshalling schedulePaymentRequest")
		return publishError
	}
	return nil
}

func insertTaskExecutionRequestRecordInDb(s *dbModels.TaskSchedule) (*dbModels.TaskExecution, pgx.Tx, error) {
	var scheduledTaskRepository repository.TaskScheduleRepository = repository.TaskScheduleRepository{}

	taskExecution, conversionError := createTaskExecutionDbModel(s)
	if conversionError != nil {
		log.Error(context.Background(), conversionError, "Error converting task schedule to task execution")
	}

	log.Info(context.Background(), "TaskExecution: %v", taskExecution)

	transaction, err := scheduledTaskRepository.InsertTaskExecutionRequestInDb(taskExecution)
	if err != nil {
		log.Error(context.Background(), err, "Error inserting task execution request in db")
		return nil, nil, err
	}
	return taskExecution, transaction, nil
}

func createTaskExecutionDbModel(taskSchedule *dbModels.TaskSchedule) (*dbModels.TaskExecution, error) {

	taskExecutionDate, dateParseError := formulateExecutionDate(taskSchedule)
	if dateParseError != nil {
		return nil, dateParseError
	}

	return &dbModels.TaskExecution{
		TaskId:              taskSchedule.TaskId,
		ExecutionDate:       taskExecutionDate,
		ExecutionParameters: taskSchedule.ExecutionParameters.ExecutionParameters,
		StartTime:           time.Now(),
		Status:              enums.TaskStaged,
	}, nil
}

func formulateExecutionDate(s *dbModels.TaskSchedule) (time.Time, error) {
	day := s.ExecutionParameters.Day
	if day == "DD" { //TODO - What happens if its DD-1 or DD-2 etc
		day = fmt.Sprintf("%02d", time.Now().Day())
	}

	month := s.ExecutionParameters.Month
	if month == "MM" {
		month = fmt.Sprintf("%02d", time.Now().Month())
	}

	year := s.ExecutionParameters.Year
	if year == "YYYY" {
		year = fmt.Sprintf("%04d", time.Now().Year())
	}

	dateString := fmt.Sprintf("%s-%s-%s", year, month, day)

	date, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		log.Info(context.Background(), err)
	}

	return date, nil
}

func areTaskDependenciesMet(s *dbModels.TaskSchedule) (*dbModels.ScheduledTask, error) {

	var scheduledTaskRepository repository.ScheduledTaskRepositoryInterface = repository.ScheduledTaskRepository{}

	scheduledTaskDbRecord, err := scheduledTaskRepository.AreTaskDependenciesMet(s.TaskId, s.NextRun)

	if err != nil {
		log.Error(context.Background(), err, "Error getting scheduled task from db: %v", err)
	}
	return scheduledTaskDbRecord, nil
}

func getTaskSchedulesFromDb() ([]*dbModels.TaskSchedule, error) {

	var taskScheduleRepository repository.TaskScheduleRepositoryInterface = repository.TaskScheduleRepository{}
	schedules, err := taskScheduleRepository.GetDueSchedules()

	if err != nil {
		log.Error(context.Background(), err, "Error getting schedule from db: %v", err)
		return nil, err
	}

	return schedules, nil
}

// getTimeFromCronInEST extracts the scheduled time from a cron expression and converts it to the EST timezone.
// It takes a TaskSchedule object as input and returns the scheduled time, current time in EST, and the next run time from the database in EST.
// The function parses the cron expression, calculates the next scheduled time, and converts it to the EST timezone.
// It also retrieves the current time in EST and the next run time from the database in EST.
// The function returns the scheduled time, current time in EST, next run time from the database in EST, and any error encountered during the process.

func getAllTimesInESTForComparison(eachTaskSchedule *dbModels.TaskSchedule) (time.Time, time.Time, time.Time, error) {

	parsedCronSchedule, err := cron.ParseStandard(eachTaskSchedule.Schedule)
	if err != nil {
		log.Error(context.Background(), err, "Error parsing cron expression: %v", eachTaskSchedule.Schedule)
	}

	cronScheduleNextValue := parsedCronSchedule.Next(time.Now())
	log.Info(context.Background(), "cronScheduleNextValue value is: %v", cronScheduleNextValue)

	// Extract the hour and minute from the next scheduled time
	cronScheduleHour := cronScheduleNextValue.Hour()
	cronScheduleMinute := cronScheduleNextValue.Minute()

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Error(context.Background(), err, "Error loading EST location")
	}

	// Get the current date time in EST
	currentTime := time.Now().In(location)
	log.Info(context.Background(), "currentTime value in EST is: %v", currentTime)

	// TODO - Okay for now. Revisit later.
	// Create a new time with the current date, hour, and minute in the EST timezone
	jobScheduleTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), cronScheduleHour, cronScheduleMinute, 0, 0, location)
	log.Info(context.Background(), "The exact job scheduled time in EST: %v", jobScheduleTime)

	nextRunFromDB := eachTaskSchedule.NextRun

	// Extract date components from nextRun fetched from database
	year, month, day := nextRunFromDB.Date()

	// Extract time components from nextRun fetched from database
	nextRunHour := nextRunFromDB.Hour()
	nextRunMinute := nextRunFromDB.Minute()

	// Convert year, month, day, hour, and minute to EST time
	nextRunFromDBInEST := time.Date(year, month, day, nextRunHour, nextRunMinute, 0, 0, location)

	return jobScheduleTime, currentTime, nextRunFromDBInEST, err
}
