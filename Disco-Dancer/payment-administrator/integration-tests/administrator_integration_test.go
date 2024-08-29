//go:build integration

package integrationtest

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/messaging"

	"geico.visualstudio.com/Billing/plutus/payment-administrator-worker/handlers"
	executorHandlers "geico.visualstudio.com/Billing/plutus/payment-executor-worker/handlers"

	"github.com/geico-private/pv-bil-frameworks/config"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

var (
	testConfig                  config.AppConfiguration
	executorTestConfig          config.AppConfiguration
	PaymentResponseTopic        string
	paymentId                   int64 = 1
	brokers                     []string
	executePaymentRequestReader *kafka.Reader
	paymentEventsReader         *kafka.Reader
	paymentResponsesReader      *kafka.Reader
	taskResponsesReader         *kafka.Reader
	executeTaskRequestTopic     string
)

func init() {
	var err error

	testConfig = config.NewConfigBuilder().
		AddJsonFile("../worker/config/appsettings.json").
		AddJsonFile("../worker/config/secrets.json").Build()

	fmt.Println("Initializing administrator integration test")

	database.Init(testConfig)

	brokers = []string{testConfig.GetString("PaymentPlatform.Kafka.Brokers", "")}
	consumerName := "dv.paymentplatform.groups.administrator"
	executorConsumerName := "dv.paymentplatform.groups.executor"

	err = messaging.InitKafka(brokers, consumerName)
	if err != nil {
		panic(err)
	}

	executeTaskRequestTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	messaging.KafkaSubscribe(executeTaskRequestTopic, handlers.ProcessTaskRequestHandler)

	handlers.PaymentResponseTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.PaymentResponses", "")
	paymentResponsesReader = commonFunctions.NewTestKafkaReader(brokers, consumerName, handlers.PaymentResponseTopic)

	handlers.ExecuteTaskResponseTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskResponses", "")
	taskResponsesReader = commonFunctions.NewTestKafkaReader(brokers, consumerName, handlers.ExecuteTaskResponseTopic)

	handlers.DataHubPaymentEventsTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.PaymentEvents", "")
	paymentEventsReader = commonFunctions.NewTestKafkaReader(brokers, consumerName, handlers.DataHubPaymentEventsTopic)

	executorTestConfig = config.NewConfigBuilder().
		AddJsonFile("../../payment-executor/worker/config/appsettings.json").
		AddJsonFile("../../payment-executor/worker/config/secrets.json").Build()
	var executePaymentRequestTopic = executorTestConfig.GetString("PaymentPlatform.Kafka.Topics.ExecutePaymentRequests", "")
	executePaymentRequestReader = commonFunctions.NewTestKafkaReader(brokers, executorConsumerName, executePaymentRequestTopic)
}

func TestExecutePaymentResponses_SuccessMessage(t *testing.T) {
	// Test case - ITG045
	// 1.Setup the Data as per the prerequisite for the test case.
	// 2.Publish the completed status for executepaymentresponses, as it is prerequisite for this test case
	// 3.Call ProcessExecutePaymentResponse handler that will read the messages from the above topic and update the respective tables - incoming_payment_request & payment
	// 4.The expected results for this test case is to verify if the status for both the tables are Completed.
	err := SeedDataToPostgresTables(truncate_Payment, truncate_PaymentRequest, insert_PaymentRequest, insert_Payment)
	require.NoError(t, err)

	executePaymentResponseTopic := testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecutePaymentResponses", "")
	messaging.KafkaSubscribe(executePaymentResponseTopic, handlers.ExecutePaymentResponseHandler)

	// publish a test message to executePaymentResponseTopic
	PublishMessageToTopic(executePaymentResponseTopic, testConfig, "", "", "")

	var paymentResponse commonMessagingModels.ExecutePaymentResponse
	err = commonFunctions.ProcessOneMesssageFromReader(paymentResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		err := json.Unmarshal([]byte(*message.Body), &paymentResponse)
		require.NoError(t, err)
		require.EqualValues(t, paymentResponse.Status, enums.Completed)
		return nil
	})
	require.NoError(t, err)

	// check the database status
	paymentStatus, incommingRequestStatus, err := VerifyPaymentRequestStatus(paymentId)
	require.NoError(t, err)
	require.EqualValues(t, enums.Completed, paymentStatus)
	require.EqualValues(t, enums.Completed, incommingRequestStatus)

}

func TestInvalidpaymentMethodType(t *testing.T) {
	// Test case - ITG028
	// 1.Publish the request message to – ExecuteTaskRequests topic with invalid payment method type, as it is prerequisite for this test case
	// 2.Call ProcessPaymentTaskHandler that will read the request messages from the above topic and validate the request
	// 3.The expected results for this test case is to verify if the payment method type is valid or not.
	executeTaskRequestTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	PublishMessageToTopic(executeTaskRequestTopic, testConfig, "badPaymentMethodType", "iaa", "oneTime") // invalid paymentMethodType

	var taskResponse commonMessagingModels.ExecuteTaskResponse
	err := commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, "errored")
		require.EqualValues(t, taskResponse.ErrorDetails.Message, "invalid input payment method type 'badPaymentMethodType'")
		return nil
	})
	require.NoError(t, err)
}

func TestInvalidPaymentRequestType(t *testing.T) {
	// Test case - ITG029
	// 1.Publish the request message to – ExecuteTaskRequests topic with payment request type, as it is prerequisite for this test case
	// 2.Call ProcessPaymentTaskHandler that will read the request messages from the above topic and validate the request
	// 3.The expected results for this test case are to verify if the payment request type is valid or not.
	executeTaskRequestTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	PublishMessageToTopic(executeTaskRequestTopic, testConfig, "ach", "badPaymentRequestType", "oneTime") // invalid PaymentRequestType

	var taskResponse commonMessagingModels.ExecuteTaskResponse
	err := commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, "errored")
		require.EqualValues(t, taskResponse.ErrorDetails.Message, "invalid input payment request type 'badPaymentRequestType'")
		return nil
	})
	require.NoError(t, err)
}

func TestInvalidPaymentFrequency(t *testing.T) {
	// Test case - ITG030
	// 1.Publish the request message to – ExecuteTaskRequests topic with invalid payment frequency , as it is prerequisite for this test case
	// 2.Call ProcessPaymentTaskHandler that will read the request messages from the above topic and validate the request
	// 3.The expected results for this test case are to verify if the payment frequency type is valid or not.
	executeTaskRequestTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	PublishMessageToTopic(executeTaskRequestTopic, testConfig, "ach", "iaa", "badPaymentFrequency") // invalid PaymentFrequency

	var taskResponse commonMessagingModels.ExecuteTaskResponse
	err := commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, "errored")
		require.EqualValues(t, taskResponse.ErrorDetails.Message, "invalid input payment frequency 'badPaymentFrequency'")
		return nil
	})
	require.NoError(t, err)
}

func TestProcessPayments_ErrorsIfNoPaymentPreferences(t *testing.T) {
	// ITG031
	// 1. Ininitalizes the DB with a payment request but no payment preferences.
	// 2. Publishes a message to execute a one time payment
	// 3. Expects the request to complete with no processed records and and error in the DB for the payment.
	dbErr := SeedDataToPostgresTables(truncate_Payment, truncate_PaymentRequest, insert_PaymentRequest, insert_Payment)
	require.NoError(t, dbErr)
	dbErr = SeedDataToVault(truncate_ProductDetails, truncate_PaymentMethod, truncate_PaymentPreference, insert_ProductDetails, insert_PaymentMethod, insert_PaymentPreferenceWithNoSplit)
	require.NoError(t, dbErr)

	executeTaskRequestTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	PublishMessageToTopic(executeTaskRequestTopic, testConfig, "ach", "iaa", "onetime")

	var taskResponse commonMessagingModels.ExecuteTaskResponse
	err := commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, enums.InProgress.String())
		require.EqualValues(t, taskResponse.ProcessedRecordsCount, 0)
		return nil
	})
	require.NoError(t, err)

	err = commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, enums.Completed.String())
		require.EqualValues(t, taskResponse.ProcessedRecordsCount, 0)
		return nil
	})
	require.NoError(t, err)

	// Check the database status
	paymentStatus, incommingRequestStatus, dbErr := VerifyPaymentRequestStatus(paymentId)
	require.NoError(t, dbErr)
	require.EqualValues(t, paymentStatus, enums.Errored.EnumIndex())
	require.EqualValues(t, incommingRequestStatus, enums.Errored.EnumIndex())
}

func TestProcessPayments_ErrorsIfPaymentPreferencesNotEvenlySplit(t *testing.T) {
	// ITG032
	// 1. Ininitalizes the DB with a payment request and payment preferences with a split not equal to 100.
	// 2. Publishes a message to execute a one time payment
	// 3. Expects the request to complete with no processed records and and error in the DB for the payment.
	dbErr := SeedDataToPostgresTables(truncate_Payment, truncate_PaymentRequest, insert_PaymentRequest, insert_Payment)
	require.NoError(t, dbErr)
	dbErr = SeedDataToVault(truncate_ProductDetails, truncate_PaymentMethod, truncate_PaymentPreference, insert_ProductDetails, insert_PaymentMethod, insert_PaymentPreferenceWithBadSplit)
	require.NoError(t, dbErr)

	executeTaskRequestTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	PublishMessageToTopic(executeTaskRequestTopic, testConfig, "ALL", "iaa", "onetime")

	var taskResponse commonMessagingModels.ExecuteTaskResponse
	err := commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, enums.InProgress.String())
		require.EqualValues(t, taskResponse.ProcessedRecordsCount, 0)
		return nil
	})
	require.NoError(t, err)

	err = commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, enums.Completed.String())
		require.EqualValues(t, taskResponse.ProcessedRecordsCount, 0)
		return nil
	})
	require.NoError(t, err)

	// Check the database status
	paymentStatus, incommingRequestStatus, dbErr := VerifyPaymentRequestStatus(paymentId)
	require.NoError(t, dbErr)
	require.EqualValues(t, paymentStatus, enums.Errored.EnumIndex())
	require.EqualValues(t, incommingRequestStatus, enums.Errored.EnumIndex())
}

func TestProcessPayments_SucceedsAndIsInsertedIntoExecuter(t *testing.T) {
	// ITG033
	// 1. Ininitalizes the DB with a payment request and payment preferences.
	// 2. Publishes a message to execute a one time ACH payment.
	// 3. Expects the request to complete with a processed record and a in progress status in the DB.
	// 4. Runs execute payment request in the executor based on the request made by the administrator.
	// 5. Expects the execution request to be entered into the executor DB with the correct values.
	handlers.ExecutePaymentRequestTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecutePaymentRequests", "")
	executorHandlers.DataHubPaymentEventsTopic = executorTestConfig.GetString("PaymentPlatform.Kafka.Topics.PaymentEvents", "")
	dbErr := SeedDataToPostgresTables(truncate_Payment, truncate_PaymentRequest, insert_PaymentRequest, insert_Payment)
	require.NoError(t, dbErr)
	dbErr = SeedDataToVault(truncate_ProductDetails, truncate_PaymentMethod, truncate_PaymentPreference, insert_ProductDetails, insert_PaymentMethod, insert_PaymentPreference)
	require.NoError(t, dbErr)

	executeTaskRequestTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	PublishMessageToTopic(executeTaskRequestTopic, testConfig, "ACH", "iaa", "onetime")

	var taskResponse commonMessagingModels.ExecuteTaskResponse
	err := commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, enums.InProgress.String())
		require.EqualValues(t, taskResponse.ProcessedRecordsCount, 0)
		return nil
	})
	require.NoError(t, err)

	err = commonFunctions.ProcessOneMesssageFromReader(taskResponsesReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		require.NoError(t, err)
		require.EqualValues(t, taskResponse.Status, enums.Completed.String())
		require.EqualValues(t, taskResponse.ProcessedRecordsCount, 1)
		return nil
	})
	require.NoError(t, err)

	// Check the database status
	paymentStatus, incommingRequestStatus, dbErr := VerifyPaymentRequestStatus(paymentId)
	require.NoError(t, dbErr)
	require.EqualValues(t, paymentStatus, enums.InProgress.EnumIndex())
	require.EqualValues(t, incommingRequestStatus, enums.InProgress.EnumIndex())

	// Read payment event sent by administrator
	err = commonFunctions.ProcessOneMesssageFromReader(paymentEventsReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		var paymentEvent commonMessagingModels.PaymentEvent
		err := json.Unmarshal([]byte(*message.Body), &paymentEvent)
		require.NoError(t, err)
		require.EqualValues(t, paymentEvent.PaymentRequestType, enums.InsuranceAutoAuctions)
		require.EqualValues(t, paymentEvent.EventType, enums.SentByAdminstrator)
		return nil
	})
	require.NoError(t, err)

	// Switching the DB context to use the executor DB.
	oldDbContext := database.GetDbContext()
	database.Init(executorTestConfig)
	newPool := database.NewPgxPool()
	database.SetDbContext(database.DbContext{Database: newPool})
	err = SeedDataToPostgresTables(truncate_ExecutionRequest)
	require.NoError(t, err)

	err = commonFunctions.ProcessOneMesssageFromReader(executePaymentRequestReader, executorHandlers.ExecutePaymentRequestHandler)
	require.NoError(t, err)

	// Read payment event sent by executor
	err = commonFunctions.ProcessOneMesssageFromReader(paymentEventsReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		var paymentEvent commonMessagingModels.PaymentEvent
		err := json.Unmarshal([]byte(*message.Body), &paymentEvent)
		require.NoError(t, err)
		require.EqualValues(t, paymentEvent.PaymentRequestType, enums.InsuranceAutoAuctions)
		require.EqualValues(t, paymentEvent.EventType, enums.ReceivedByExecutor)
		return nil
	})
	require.NoError(t, err)

	// Check the executionRequest table
	executionRequest, err := GetExecutionRequestFromDB(paymentId)
	require.NoError(t, err)
	require.EqualValues(t, executionRequest.ExecutionRequestId, 1)
	require.EqualValues(t, executionRequest.PaymentId, paymentId)
	require.EqualValues(t, executionRequest.Amount, float32(20.16))
	require.EqualValues(t, executionRequest.Last4AccountIdentifier, "1234")
	require.EqualValues(t, executionRequest.PaymentRequestType, enums.InsuranceAutoAuctions)
	require.EqualValues(t, executionRequest.PaymentMethodType, enums.ACH)

	// Closes the executor db connection and switches back to the admin db connection.
	database.GetDbContext().Database.Close()
	database.SetDbContext(*oldDbContext)
}
