//go:build integration

package integrationtest

import (
	"context"
	"encoding/json"
	"testing"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/crypto"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	filmapiclient "geico.visualstudio.com/Billing/plutus/filmapi-client"
	"geico.visualstudio.com/Billing/plutus/messaging"
	AchOnetime "geico.visualstudio.com/Billing/plutus/payment-executor-common/settlement/outbound/ach-onetime"

	"geico.visualstudio.com/Billing/plutus/payment-executor-worker/handlers"
	"github.com/geico-private/pv-bil-frameworks/config"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	expectedAchOnetimeFileName = "GPP.ACH.NACHA.TXT"
)

var (
	testConfig                              config.AppConfiguration
	brokers                                 []string
	executeTaskRequestTopic                 string
	taskRequestsReader                      *kafka.Reader
	taskResponsesReader                     *kafka.Reader
	inboundFileFileProcessingCompletedTopic string
	inboundFileProcessReader                *kafka.Reader
	executePaymentResponseTopic             string
	executePaymentResponseReader            *kafka.Reader
	consolidatedResponseTopic               string
	consoliDatedResponseReader              *kafka.Reader
	inboundFileProcessRecordTopic           string
	inboundFileProcessRecordFeedbackTopic   string
	inboundFileProcessRecordReader          *kafka.Reader
	inboundFileProcessRecordFeedbackReader  *kafka.Reader
)

func init() {
	testConfig = config.NewConfigBuilder().
		AddJsonFile("../worker/config/appsettings.json").
		AddJsonFile("../worker/config/secrets.json").Build()

	brokers = []string{testConfig.GetString("PaymentPlatform.Kafka.Brokers", "")}
	consumerName := "dv.paymentplatform.groups.executor"
	err := messaging.InitKafka(brokers, consumerName)

	if err != nil {
		panic(err)
	}

	database.Init(testConfig)

	volatageSideCarAddress := testConfig.GetString("Voltage.Sidecar.Address", "")
	crypto.Init(volatageSideCarAddress)

	url := testConfig.GetString("Filmapi.Url", "")
	authorization := testConfig.GetString("Filmapi.Authorization", "")
	timeout := testConfig.GetInt("ExternalApi.Timeout", 0)
	retryCount := testConfig.GetInt("ExternalApi.Retrycount", 0)

	filmapiclient.Init(url, authorization, timeout, retryCount, nil)
	commonFunctions.SetConfigHandler(testConfig)

	executeTaskRequestTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	inboundFileFileProcessingCompletedTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.InboundFileFileProcessingCompleted", "")
	handlers.ExecuteTaskResponseTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskResponses", "")
	executePaymentResponseTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecutePaymentResponses", "")
	consolidatedResponseTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ConsolidatedPayments", "")
	inboundFileProcessRecordTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.InboundFileProcessRecord", "")
	inboundFileProcessRecordFeedbackTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.InboundFileProcessRecordFeedback", "")

	taskRequestsReader = commonFunctions.NewTestKafkaReader(brokers, consumerName, executeTaskRequestTopic)
	taskResponsesReader = commonFunctions.NewTestKafkaReader(brokers, consumerName, handlers.ExecuteTaskResponseTopic)
	inboundFileProcessReader = commonFunctions.NewTestKafkaReader(brokers, consumerName, inboundFileFileProcessingCompletedTopic)
	executePaymentResponseReader = commonFunctions.NewTestKafkaReader(brokers, consumerName, executePaymentResponseTopic)
	consoliDatedResponseReader = commonFunctions.NewTestKafkaReader(brokers, consumerName, consolidatedResponseTopic)
	inboundFileProcessRecordReader = commonFunctions.NewTestKafkaReader(brokers, consumerName, inboundFileProcessRecordTopic)
	inboundFileProcessRecordFeedbackReader = commonFunctions.NewTestKafkaReader(brokers, consumerName, inboundFileProcessRecordFeedbackTopic)
}

func TestUpdateSettlementIdentifier(t *testing.T) {
	// Test case - ITG040
	// 1.Setup the Data as per the prerequisite for the test case.
	// 2.Publish the message for ExecuteTaskRequests, as it is prerequisite for this test case
	// 3.Subscribe ExecuteTaskResponsesTopic to receive message
	// 4.The expected results for this test case is to verify the updated Settlement Identifier in the database
	err := SeedDataToPostgresTables(truncate__consolidated_request, truncate_execution_request, GetInsertConsolidatedRequestQueries(1, 1), GetInsertExecutionRequestQueries(1, 1))
	require.NoError(t, err)
	// Publish a test message to settlePaymentsTaskRequestsTopic
	t.Log("Publishing message to topic: ", executeTaskRequestTopic)
	PublishMessageToTopic(executeTaskRequestTopic, testConfig, "ach", "all", "all")

	err = commonFunctions.ProcessOneMesssageFromReader(taskRequestsReader, handlers.PaymentSettlementTaskHandler)
	require.NoError(t, err)

	var taskResponse commonMessagingModels.ExecuteTaskResponse
	err = commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		// validate the first message is inprogress stuass
		err = json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		assert.EqualValues(t, enums.TaskInprogress.String(), taskResponse.Status)
		return nil
	})
	require.NoError(t, err)

	err = commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		// validate the second message is completed status
		err = json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		assert.EqualValues(t, enums.TaskCompleted.String(), taskResponse.Status)
		return nil
	})
	// check the database status
	var consolidatedId int64 = 12346
	settlementIdentifier, err := GetSettlementIdentifier(consolidatedId)
	require.NoError(t, err)
	assert.EqualValues(t, AchOnetime.GetIndividualId(consolidatedId, 15, "C"), settlementIdentifier)
}

func TestUpdateExecutionRequestToComplete(t *testing.T) {
	// Test case - ITG043
	// 1.Setup the Data as per the prerequisite for the test case.
	// 2.Publish the message for InboundFileFileProcessingCompleted, as it is prerequisite for this test case
	// 3.Subscribe ExecutePaymentResponses to receive message
	// 4.The expected results for this test case is to verify if the status for both the table and response message are Completed.

	err := SeedDataToPostgresTables(truncate_execution_request, GetInsertExecutionRequestQueries(4, 3))
	require.NoError(t, err)

	inboundFileFileProcessingCompletedTopic := testConfig.GetString("PaymentPlatform.Kafka.Topics.InboundFileFileProcessingCompleted", "")

	// Publish a test message to inboundFileFileProcessingCompletedTopic
	PublishMessageToTopic(inboundFileFileProcessingCompletedTopic, testConfig, "", "", "")
	err = commonFunctions.ProcessOneMesssageFromReader(inboundFileProcessReader, handlers.FileProcessingCompletedHandler)
	require.NoError(t, err)

	for i := 1; i <= 3; i++ {
		err = commonFunctions.ProcessOneMesssageFromReader(executePaymentResponseReader, func(ctx context.Context, message *kafkamessaging.Message) error {
			modifiedBody := GetJsonDate(message)
			var completedPaymentResponse commonMessagingModels.ExecutePaymentResponse
			err = json.Unmarshal([]byte(modifiedBody), &completedPaymentResponse)
			require.NoError(t, err)
			assert.EqualValues(t, enums.Completed.EnumIndex(), completedPaymentResponse.Status)

			// Check the database status
			status, err := GetExecutionRequestStatus(completedPaymentResponse.PaymentId)
			require.NoError(t, err)
			assert.EqualValues(t, enums.Completed.EnumIndex(), status)
			return nil
		})
		require.NoError(t, err)
	}
}

func TestInvalidpaymentMethodType(t *testing.T) {
	// Test case - ITG036
	// 1.Publish the request message to – ExecuteTaskRequests topic with invalid payment method type, as it is prerequisite for this test case
	// 2.Call ProcessPaymentTaskHandler that will read the request messages from the above topic and validate the request
	// 3.The expected results for this test case is to verify if the payment method type is valid or not.

	PublishMessageToTopic(executeTaskRequestTopic, testConfig, "invalid", "all", "all") // invalid paymentMethodType
	err := commonFunctions.ProcessOneMesssageFromReader(taskRequestsReader, handlers.PaymentSettlementTaskHandler)
	require.ErrorContains(t, err, "invalid input payment method type 'invalid'")

	err = commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		var taskResponse commonMessagingModels.ExecuteTaskResponse
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, "errored")
		require.EqualValues(t, taskResponse.ErrorDetails.Message, "invalid input payment method type 'invalid'")
		return nil
	})
	require.NoError(t, err)
}
func TestInvalidPaymentRequestType(t *testing.T) {
	// Test case - ITG037
	// 1.Publish the request message to – ExecuteTaskRequests topic with payment request type, as it is prerequisite for this test case
	// 2.Call ProcessPaymentTaskHandler that will read the request messages from the above topic and validate the request
	// 3.The expected results for this test case are to verify if the payment request type is valid or not.
	PublishMessageToTopic(executeTaskRequestTopic, testConfig, "ach", "invalid", "oneTime") // invalid PaymentRequestType
	err := commonFunctions.ProcessOneMesssageFromReader(taskRequestsReader, handlers.PaymentSettlementTaskHandler)
	require.ErrorContains(t, err, "invalid input payment request type 'invalid'")

	err = commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		var taskResponse commonMessagingModels.ExecuteTaskResponse
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, "errored")
		require.EqualValues(t, taskResponse.ErrorDetails.Message, "invalid input payment request type 'invalid'")
		return nil
	})
	require.NoError(t, err)
}

func TestInvalidPaymentFrequency(t *testing.T) {
	// Test case - ITG038
	// 1.Publish the request message to – ExecuteTaskRequests topic with invalid payment frequency , as it is prerequisite for this test case
	// 2.Call ProcessPaymentTaskHandler that will read the request messages from the above topic and validate the request
	// 3.The expected results for this test case are to verify if the payment frequency type is valid or not.
	PublishMessageToTopic(executeTaskRequestTopic, testConfig, "ach", "all", "invalid") // invalid PaymentFrequency
	err := commonFunctions.ProcessOneMesssageFromReader(taskRequestsReader, handlers.PaymentSettlementTaskHandler)
	require.ErrorContains(t, err, "invalid input payment frequency 'invalid'")

	err = commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		var taskResponse commonMessagingModels.ExecuteTaskResponse
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, "errored")
		require.EqualValues(t, taskResponse.ErrorDetails.Message, "invalid input payment frequency 'invalid'")
		return nil
	})
	require.NoError(t, err)
}

func TestUpdateConsolidatedRequestToComplete(t *testing.T) {
	// Test case - ITG044
	// 1.Setup the Data as per the prerequisite for the test case.
	// 2.Publish the message for InboundFileFileProcessingCompleted, as it is prerequisite for this test case
	// 3.Subscribe ConsolidatedPayments to receive message
	// 4.The expected results for this test case is to verify if the status for both the table and response message are Completed.
	err := SeedDataToPostgresTables(truncate__consolidated_request, truncate_execution_request, GetInsertConsolidatedRequestQueries(4, 1))
	require.NoError(t, err)

	// Publish a test message to inboundFileFileProcessingCompletedTopic
	PublishMessageToTopic(inboundFileFileProcessingCompletedTopic, testConfig, "", "", "")

	err = commonFunctions.ProcessOneMesssageFromReader(inboundFileProcessReader, handlers.FileProcessingCompletedHandler)
	require.NoError(t, err)

	err = commonFunctions.ProcessOneMesssageFromReader(consoliDatedResponseReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		var completedPaymentResponse commonMessagingModels.ConsolidatedExecutePaymentResponse
		err := json.Unmarshal([]byte(*message.Body), &completedPaymentResponse)
		require.NoError(t, err)
		require.EqualValues(t, enums.Completed.String(), completedPaymentResponse.Status)

		// Check the database status
		status, err := GetConsolidatedStatus(completedPaymentResponse.ConsolidatedId)
		require.NoError(t, err)
		require.EqualValues(t, enums.Completed.EnumIndex(), status)
		return nil
	})
	require.NoError(t, err)
}
