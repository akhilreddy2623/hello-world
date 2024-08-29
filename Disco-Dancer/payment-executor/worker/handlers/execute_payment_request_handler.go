package handlers

import (
	"context"
	"encoding/json"
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/repository"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
)

var log = logging.GetLogger("payment-executor-handlers")
var DataHubPaymentEventsTopic string

var ExecutePaymentRequestHandler = func(ctx context.Context, message *kafkamessaging.Message) error {
	log.Info(ctx, "Received message in paymentplatform.internal.executepaymentrequests topic: '%s'", *message.Body)

	var executePaymentRequest commonMessagingModels.ProcessPaymentRequest
	var executePaymentRequestRepository repository.ExecutionRequestRepositoryInterface = repository.ExecutionRequestRepository{}

	err := json.Unmarshal([]byte(*message.Body), &executePaymentRequest)
	if err != nil {
		log.Error(ctx, err, "Error in unmarshalling execute payment request message")
		return err
	}

	err = executePaymentRequestProcess(ctx, executePaymentRequest, executePaymentRequestRepository)

	if err != nil {
		log.Error(context.Background(), err, "error processing execute payment request")
		return err
	}

	return nil
}

func executePaymentRequestProcess(
	ctx context.Context,
	request commonMessagingModels.ProcessPaymentRequest,
	repository repository.ExecutionRequestRepositoryInterface) error {

	log.Info(ctx, "Processing execute payment request")

	//TODO - Validation (Not Day 1) - Because the data being in queue is already pre-validated data from the vault
	err := repository.ExecuteRequests(request)
	// Publish Payment Event- ReceivedByExecutor  to DataHub
	if err == nil {
		publishPaymentEventToDataHub(ctx, request)
	}

	return err
}

func publishPaymentEventToDataHub(ctx context.Context, request commonMessagingModels.ProcessPaymentRequest) {
	processPaymentEvent, err := createPaymentEventJson(request)
	if err != nil {
		log.Error(ctx, err, "error before publishing the payment event message to datahub, payment id '%d'", request.PaymentId)
	}

	err = messaging.KafkaPublish(DataHubPaymentEventsTopic, *processPaymentEvent)
	if err != nil {
		log.Error(ctx, err, "error while publishing the payment event message to datahub, payment id '%d'", request.PaymentId)
	}
}

func createPaymentEventJson(processPaymentRequest commonMessagingModels.ProcessPaymentRequest) (*string, error) {
	paymentEvent := commonMessagingModels.PaymentEvent{
		Version:            1,
		PaymentId:          processPaymentRequest.PaymentId,
		Amount:             processPaymentRequest.Amount,
		PaymentDate:        time.Time(processPaymentRequest.PaymentDate),
		PaymentRequestType: enums.GetPaymentRequestTypeEnum(processPaymentRequest.PaymentRequestType),
		PaymentMethodType:  enums.GetPaymentMethodTypeEnumFromString(processPaymentRequest.PaymentMethodType),
		EventType:          enums.ReceivedByExecutor,
		EventDateTime:      time.Now(),
	}
	paymentEventJson, err := json.Marshal(paymentEvent)
	if err != nil {
		log.Error(context.Background(), err, "unable to marshal processPaymentRequest")
		return nil, err
	}

	paymentEventJsonStr := string(paymentEventJson)
	return &paymentEventJsonStr, nil
}
