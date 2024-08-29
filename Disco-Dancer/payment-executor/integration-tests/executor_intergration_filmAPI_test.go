//go:build integration && azureEndpoint

package integrationtest

import (
	"context"
	"encoding/json"
	"testing"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/enums"
	filmapiclient "geico.visualstudio.com/Billing/plutus/filmapi-client"
	inbound_fileprocessor "geico.visualstudio.com/Billing/plutus/inbound-fileprocessor"
	settlement "geico.visualstudio.com/Billing/plutus/payment-executor-common/settlement/outbound"
	"geico.visualstudio.com/Billing/plutus/payment-executor-worker/handlers"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
	"github.com/stretchr/testify/require"
)

// var (
// 	testConfig              config.AppConfiguration
// 	brokers                 []string
// 	executeTaskRequestTopic string
// 	taskRequestsReader      *kafka.Reader
// 	taskResponsesReader     *kafka.Reader
// )

// const (
// 	expectedAchOnetimeFileName = "GPP.ACH.NACHA.TXT"
// )

func TestFileGenerationAndFileUpload(t *testing.T) {
	// Test case - ITG041
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Publish the message for ExecuteTaskRequests, as it is prerequisite for this test case
	// 3. Subscribe ExecuteTaskResponsesTopic to receive message
	// 4. The expected results for this test case is to verify the File is created in the blob
	err := SeedDataToPostgresTables(truncate__consolidated_request, truncate_execution_request, GetInsertConsolidatedRequestQueries(1, 5), GetInsertExecutionRequestQueries(1, 1))
	require.NoError(t, err)

	handlers.DataHubPaymentEventsTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.PaymentEvents", "")
	settlement.Outboundfileprocessingpath = testConfig.GetString("TestOutboundFileProcessingPath.InsuranceAutoAuctions", "")

	// Publish a test message to settlePaymentsTaskRequestsTopic
	PublishMessageToTopic(executeTaskRequestTopic, testConfig, "ach", "all", "all")
	err = commonFunctions.ProcessOneMesssageFromReader(taskRequestsReader, handlers.PaymentSettlementTaskHandler)
	require.NoError(t, err)
	// messaging.KafkaSubscribe(settlePaymentsTaskRequestsTopic, handlers.PaymentSettlementTaskHandler)

	// Wait for the "completed" message
	for taskCompleted := false; !taskCompleted; {
		err = commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
			var taskResponse commonMessagingModels.ExecuteTaskResponse
			err := json.Unmarshal([]byte(*message.Body), &taskResponse)
			require.NoError(t, err)
			if taskResponse.Status == enums.TaskCompleted.String() {
				taskCompleted = true
			}
			return nil
		})
		require.NoError(t, err)
	}

	// Verify the file upload
	fileNames, err := filmapiclient.GetFilesInFolder(settlement.Outboundfileprocessingpath)
	require.NoError(t, err)

	expectedFileName := settlement.Outboundfileprocessingpath + expectedAchOnetimeFileName

	fileFound := false
	for _, fileName := range fileNames {
		if fileName == expectedFileName {
			fileFound = true
			break
		}
	}

	if fileFound {
		// Delete completed file.
		err := filmapiclient.DeleteFile(expectedFileName)
		require.NoError(t, err)
	}

	require.True(t, fileFound, "Expected file was not found in the directory")
}

func TestFileDeduplicationAndRecordDeduplication(t *testing.T) {
	// Test case - ITG042
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Start InboundFileFileProcessing.
	// 3. Subscribe FileProcessingCompleted to receive a message.
	// 4. The expected result for this test case is to verify if the status for both the table and response message are Completed.

	err := SeedDataToPostgresTables(truncate__consolidated_request, truncate_execution_request, truncate__file_deduplication, truncate_record_deduplication, GetInsertConsolidatedRequestQueries(4, 1))
	require.NoError(t, err)

	pollingInterval := testConfig.GetString("TestInboundFileProcessing.AchOnetimeAck.Polling.Interval", "")
	folderLocation := testConfig.GetString("TestInboundFileProcessing.AchOnetimeAck.Folder.Location", "")
	archiveLocation := testConfig.GetString("TestInboundFileProcessing.AchOnetimeAck.Archive.Location", "")

	// Upload the test file
	fileLocation := folderLocation + expectedAchOnetimeFileName
	fileContent := "101 02100002191577286012404160835E094101JP MORGAN CHASE        GEICO INSURANCE COMPANY00009953\n9000001001093000109268716030595000000000000002676466949"

	err = filmapiclient.UploadFile(fileLocation, fileContent)
	require.NoError(t, err, "error uploading file using film api at path: %s", folderLocation)

	// Start the file processor
	fileProcessorInput := inbound_fileprocessor.FileProcessorInput{
		PollingDurationInMins:  pollingInterval,
		FolderLocation:         folderLocation,
		BusinessFileType:       "achonetimeack",
		ArchiveFolderLocation:  archiveLocation,
		ProcessRecordTopicName: inboundFileProcessRecordTopic,
	}
	go inbound_fileprocessor.Process(fileProcessorInput)

	for i := 0; i < 2; i++ {
		// Process a single message from the ProcessRecord topic
		err = commonFunctions.ProcessOneMesssageFromReader(inboundFileProcessRecordReader, handlers.InboundFileProcessRecordHandler)
		require.NoError(t, err)

		// Process a single message from the ProcessRecordFeedBack topic
		err = commonFunctions.ProcessOneMesssageFromReader(inboundFileProcessRecordFeedbackReader, inbound_fileprocessor.ProcessRecordFeedbackHandler)
		require.NoError(t, err)
	}

	// Process a single message from the file processing completed topic
	err = commonFunctions.ProcessOneMesssageFromReader(inboundFileProcessReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		require.NotNil(t, message)

		// Validate the "completed" message
		var completedTaskResponse inbound_fileprocessor.FileProcessingCompleted
		err := json.Unmarshal([]byte(*message.Body), &completedTaskResponse)
		require.NoError(t, err)
		require.EqualValues(t, "achonetimeack", completedTaskResponse.BusinessFileType)

		return nil
	})
	require.NoError(t, err)
}
