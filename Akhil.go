package integrationtest

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/crypto"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/messaging"
	AchOnetime "geico.visualstudio.com/Billing/plutus/payment-executor-common/settlement/outbound/ach-onetime"

	"geico.visualstudio.com/Billing/plutus/payment-executor-worker/handlers"

	"github.com/geico-private/pv-bil-frameworks/config"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
	"github.com/stretchr/testify/require"
)

var (
	testConfig           config.AppConfiguration
	PaymentResponseTopic string
)

func init() {
	var err error

	// Use the new method to get the config handler
	testConfig = commonFunctions.GetConfigHandlerWithPaths(
		[]string{"../worker/config/appsettings.json", "../worker/config/secrets.json"},
	)

	fmt.Println("Initializing executor integration test")

	database.Init(testConfig)

	volatageSideCarAddress := testConfig.GetString("Voltage.Sidecar.Address", "localhost:50051")
	crypto.Init(volatageSideCarAddress)

	err = messaging.InitKafka([]string{testConfig.GetString("PaymentPlatform.Kafka.Brokers", "")}, "dv.paymentplatform.groups.executor")
	if err != nil {
		panic(err)
	}

}

func estUpdateSettlementIdentifier(t *testing.T) {
	err := SeedDataToPostgresTables(truncate__consolidated_request, truncate_execution_request, GetInsertConsolidatedRequestQuery(1), GetInsertExecutionRequestQuery(1))
	require.NoError(t, err)

	settlePaymentsTaskRequestsTopic := testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	handlers.ExecuteTaskResponseTopic = testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskResponses", "")

	messaging.KafkaSubscribe(settlePaymentsTaskRequestsTopic, handlers.PaymentSettlementTaskHandler)

	// // waiting for the PaymentResponseTopic to received a message
	paymentResponseChan := make(chan *kafkamessaging.Message)
	go messaging.KafkaSubscribe(handlers.ExecuteTaskResponseTopic, func(ctx context.Context, message *kafkamessaging.Message) error {
		paymentResponseChan <- message
		return nil
	})

	// Channel to synchronize message reception
	completedResponseChan := make(chan *kafkamessaging.Message)

	// Subscribe to the response topic before publishing the request
	go messaging.KafkaSubscribe(handlers.ExecuteTaskResponseTopic, func(ctx context.Context, message *kafkamessaging.Message) error {
		var taskResponse commonMessagingModels.ExecuteTaskResponse
		err := json.Unmarshal([]byte(*message.Body), &taskResponse)
		if err != nil {
			return err
		}

		if taskResponse.Status == enums.TaskCompleted.String() {
			completedResponseChan <- message
		}
		return nil
	})

	// Publish a test message to settlePaymentsTaskRequestsTopic
	PublishMessageToTopic(settlePaymentsTaskRequestsTopic, testConfig)

	// Wait for the "completed" message
	completedResponse := <-completedResponseChan
	require.NotNil(t, completedResponse)

	// Validate the "completed" message
	var completedTaskResponse commonMessagingModels.ExecuteTaskResponse
	err = json.Unmarshal([]byte(*completedResponse.Body), &completedTaskResponse)
	require.NoError(t, err)
	require.EqualValues(t, enums.TaskCompleted.String(), completedTaskResponse.Status)

	// check the database status
	var consolidatedId int64 = 12345
	settlementIdentifier, err := GetSettlementIdentifier(consolidatedId)
	require.NoError(t, err)
	require.EqualValues(t, AchOnetime.GetIndividualId(consolidatedId, 15, "C"), settlementIdentifier)
}

func TestUpdateExecutionRequestToComplete(t *testing.T) {
	err := SeedDataToPostgresTables(truncate_execution_request, GetInsertExecutionRequestQuery(4))
	require.NoError(t, err)

	inboundFileFileProcessingCompletedTopic := testConfig.GetString("PaymentPlatform.Kafka.Topics.InboundFileFileProcessingCompleted", "")
	executePaymentResponseTopic := testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecutePaymentResponses", "")

	messaging.KafkaSubscribe(inboundFileFileProcessingCompletedTopic, handlers.FileProcessingCompletedHandler)

	// Channel and WaitGroup for synchronization
	paymentResponseChan := make(chan *kafkamessaging.Message, 3)
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		messaging.KafkaSubscribe(executePaymentResponseTopic, func(ctx context.Context, message *kafkamessaging.Message) error {
			paymentResponseChan <- message
			wg.Done()
			return nil
		})
	}()

	// Publish a test message to inboundFileFileProcessingCompletedTopic
	PublishMessageToTopic(inboundFileFileProcessingCompletedTopic, testConfig)

	// Wait for all messages to be processed
	wg.Wait()
	close(paymentResponseChan)

	// Validate each message
	for completedResponse := range paymentResponseChan {
		require.NotNil(t, completedResponse)

		modifiedBody := GetJsonDate(completedResponse)

		var completedPaymentResponse commonMessagingModels.ExecutePaymentResponse
		err = json.Unmarshal([]byte(modifiedBody), &completedPaymentResponse)
		require.NoError(t, err)
		require.EqualValues(t, enums.Completed.EnumIndex(), completedPaymentResponse.Status)

		// Check the database status
		status, err := GetExecutionRequestStatus(completedPaymentResponse.PaymentId)
		require.NoError(t, err)
		require.EqualValues(t, enums.Completed.EnumIndex(), status)
	}
}
