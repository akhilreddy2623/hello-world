package inbound_fileprocessor

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/database"
	filmapiclient "geico.visualstudio.com/Billing/plutus/filmapi-client"
	pglock "geico.visualstudio.com/Billing/plutus/lock"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
)

const (
	inboundFileProcessingLockId pglock.ResourceLockId = 987
	inprogress                                        = "inprogress"
	retry                                             = "retry"
	incomplete                                        = "incomplete"
	complete                                          = "complete"
	errored                                           = "errored"
)

var log = logging.GetLogger("inbound-fileprocessor")
var ProcessRecordFeedbackHandler = processRecordFeedbackHandler

func Init(processRecordFeedbackTopicName string) {
	messaging.KafkaSubscribe(processRecordFeedbackTopicName, processRecordFeedbackHandler)
}

func Process(fileProcessorInput FileProcessorInput) {
	for {
		time.Sleep(time.Duration(fileProcessorInput.PollingDurationInMins * int(time.Minute)))

		fileNames, err := filmapiclient.GetFilesInFolder(fileProcessorInput.FolderLocation)
		if err != nil {
			log.Error(context.Background(), err, "film api get files call is unsuccesful")
		} else {
			log.Info(context.Background(), "film api get files call is succesful")
		}

		if len(fileNames) > 0 {
			for _, fileName := range fileNames {
				processFile(fileProcessorInput, fileName)
			}
		} else {
			continue
		}

	}
}

func processFile(fileProcessorInput FileProcessorInput, filePath string) {
	mutex, _ := pglock.NewPSqlLockService()

	var err = mutex.RunWithLock(
		context.Background(),
		func() error {
			return processFileInMutex(fileProcessorInput, filePath)
		},
		inboundFileProcessingLockId)

	if err != nil {
		log.Error(context.Background(), err, "Error recieved from processing file using mutex, businessfiletype '%s'", fileProcessorInput.BusinessFileType)
	}
}

func processFileInMutex(fileProcessorInput FileProcessorInput, filePath string) error {

	fileRecords, err := filmapiclient.GetFileLines(filePath)
	if err != nil {
		return err
	}
	fileHash := getHashForSlice(fileRecords)

	isFileHashPresent, err := isFileHashPresent(fileHash)
	if err != nil {
		return err
	}

	isFileHashWithRetryStatusPresent, err := isFileHashWithStatusPresent(fileHash, retry)
	if err != nil {
		return err
	}

	shouldProcessFile := false
	var fileId int
	if !isFileHashPresent {
		fileId, err = addFileDeduplicationEntry(filePath, fileProcessorInput.BusinessFileType, fileHash, inprogress)
		if err != nil {
			return err
		}
		shouldProcessFile = true
	} else if isFileHashWithRetryStatusPresent {
		fileId, err = updateFileDeduplicationEntry(fileHash, inprogress)
		if err != nil {
			return err
		}
		shouldProcessFile = true
	}

	if shouldProcessFile {
		totalRecordCount := 0

		for rowNumber, recordContent := range fileRecords {
			if recordContent == "" {
				continue
			}
			totalRecordCount = rowNumber + 1
		}

		for rowNumber, recordContent := range fileRecords {
			if recordContent == "" {
				continue
			}
			rowNumber++
			processRecord(recordContent, rowNumber, fileId, totalRecordCount, filePath, fileProcessorInput)
		}
	} else {
		isFileHashWithCompleteStatusPresent, err := isFileHashWithStatusPresent(fileHash, complete)
		if err != nil {
			return err
		}
		if isFileHashWithCompleteStatusPresent {
			log.Info(context.Background(), "file with path '%s' is already completed processing, invoking cleanup routine", filePath)
			//Moves completed file to archive location.
			archiveFile(filePath, fileProcessorInput.ArchiveFolderLocation)
		} else {
			log.Info(context.Background(), "Unable to process file with path '%s' failed as file deduplication failed", filePath)
		}
	}
	return nil
}

func processRecord(recordContent string, rowNumber int, fileId int, totalRecordsCount int, filePath string, fileProcessorInput FileProcessorInput) {
	recordHash := getHashForString(recordContent)

	if !isRecordHashPresent(recordHash, fileId) {
		transaction, err := database.NewDbContext().Database.Begin(context.Background())
		if err != nil {
			log.Error(context.Background(), err, "error in starting transaction for processing record with row number '%d'", rowNumber)
		}

		recordId := addRecordDeduplicationEntry(transaction, fileId, recordHash, inprogress, rowNumber)

		recordString := getRecordStructInStringFormat(fileId, recordId, recordContent, totalRecordsCount, rowNumber, filePath, fileProcessorInput)
		err = messaging.KafkaPublish(fileProcessorInput.ProcessRecordTopicName, recordString)

		if err == nil {
			err = transaction.Commit(context.Background())
			if err != nil {
				log.Error(context.Background(), err, "error in commiting transaction for processing record with row number '%d'", rowNumber)
			}
		} else {
			err = transaction.Rollback(context.Background())
			if err != nil {
				log.Error(context.Background(), err, "error in rollbacking transaction for processing record with row number '%d'", rowNumber)
			}
		}

	} else {
		log.Info(context.Background(), "unable to process record with row number '%d' failed as record deduplication failed", rowNumber)
	}

}

var processRecordFeedbackHandler = func(ctx context.Context, message *kafkamessaging.Message) error {

	log.Info(ctx, "PaymentExecutor - Received message in paymentplatform.inboundfileprocessor.processrecordfeedback topic to process file record feedback: '%s'", message.MessageID)

	recordFeedback := RecordFeedback{}
	err := json.Unmarshal([]byte(*message.Body), &recordFeedback)
	if err != nil {
		log.Error(ctx, err, "error in unmarshalling record feedback message")
		return err
	}

	if recordFeedback.IsError {
		log.Info(context.Background(), "record id '%d' is marked as error by application", recordFeedback.RecordId)
		transaction, err := database.NewDbContext().Database.Begin(context.Background())
		defer transaction.Rollback(context.Background())
		if err != nil {
			log.Error(context.Background(), err, "error in starting transaction for processing record feedback with id '%d'", recordFeedback.RecordId)
		}

		err = updateRecordDeduplicationEntry(transaction, recordFeedback.RecordId, recordFeedback.FileId, errored)
		if err != nil {
			return err
		}
		err = updateFileDeduplicationEntryForGivenFileId(transaction, recordFeedback.FileId, incomplete)
		if err != nil {
			return err
		}

		transaction.Commit(context.Background())

	} else {
		log.Info(context.Background(), "record id '%d' is marked as success by application", recordFeedback.RecordId)
		err := updateRecordDeduplicationEntryWithNoTransaction(recordFeedback.RecordId, recordFeedback.FileId, complete)
		if err != nil {
			return err
		}

		completedRecordCount, err := getRecordCountWithSpecificStatus(recordFeedback.FileId, complete)
		if err != nil {
			return err
		}
		if completedRecordCount == recordFeedback.TotalRecordCount {
			log.Info(context.Background(), "all records for a file Id '%d', are completed processing succesfully", recordFeedback.FileId)
			transaction, err := database.NewDbContext().Database.Begin(context.Background())
			defer transaction.Rollback(context.Background())
			if err != nil {
				log.Error(context.Background(), err, "error in starting transaction for processing record feedback with id '%d'", recordFeedback.RecordId)
			}

			err = updateFileDeduplicationEntryForGivenFileId(transaction, recordFeedback.FileId, complete)
			if err != nil {
				return err
			}
			configHandler := commonFunctions.GetConfigHandler()
			var fileCompletionTopicName = configHandler.GetString("PaymentPlatform.Kafka.Topics.InboundFileFileProcessingCompleted", "")
			fileProcessingCompletedString := getFileProcessingCompletedString(recordFeedback.BusinessFileType)
			err = messaging.KafkaPublish(fileCompletionTopicName, fileProcessingCompletedString)

			if err == nil {
				err = transaction.Commit(context.Background())
				if err != nil {
					log.Error(context.Background(), err, "error in commiting transaction for processing record feedback with recordId '%d'", recordFeedback.RecordId)
				}
			} else {
				err = transaction.Rollback(context.Background())
				if err != nil {
					log.Error(context.Background(), err, "error in rollbacking transaction for processing record feedback with recordId  '%d'", recordFeedback.RecordId)
				}
			}

			//Moves completed file to archive location.
			archiveFile(recordFeedback.FilePath, recordFeedback.ArchiveFolderLocation)
		}

	}

	return nil
}

func getFileProcessingCompletedString(businessFileType string) string {
	fileProcessingCompleted := FileProcessingCompleted{
		BusinessFileType: businessFileType,
	}

	fileProcessingCompletedJson, err := json.Marshal(fileProcessingCompleted)
	if err != nil {
		log.Error(context.Background(), err, "error in marshalling fileProcessingCompleted struct to json")
	}
	return string(fileProcessingCompletedJson)
}

func getRecordStructInStringFormat(fileId int, recordId int64, recordContent string, totalRecordCount int, rowNumber int, filePath string, fileProcessorInput FileProcessorInput) string {
	record := Record{
		FileId:                fileId,
		RecordId:              recordId,
		RecordContent:         recordContent,
		BusinessFileType:      fileProcessorInput.BusinessFileType,
		TotalRecordCount:      totalRecordCount,
		FilePath:              filePath,
		ArchiveFolderLocation: fileProcessorInput.ArchiveFolderLocation,
	}

	recordJson, err := json.Marshal(record)
	if err != nil {
		log.Error(context.Background(), err, "error in marshalling file record with row number '%d'", rowNumber)
	}
	return string(recordJson)
}

func getHashForSlice(records []string) string {
	var sb strings.Builder
	for _, str := range records {
		sb.WriteString(str)
		sb.WriteString("\n")
	}
	str := sb.String()
	if len(str) > 0 {
		str = str[:len(str)-1]
	}
	return getHashForString(str)
}

func getHashForString(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func archiveFile(file string, archiveDir string) {

	now := time.Now()

	dateStr := now.Format("20060102") //YYYYMMDD
	timeStr := now.Format("150405")   //HHMMSS

	_, fileName := filepath.Split(file)

	newFileName := filepath.Join(archiveDir, dateStr, timeStr, fileName)

	err := filmapiclient.MoveFile(file, newFileName)

	if err != nil {
		log.Error(context.Background(), err, "Error moving the file to archive location")
	}
}
