package handlers

import (
	"context"
	"encoding/json"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-common/repository"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
)

var PaymentResponseTopic string

//go:generate mockery --name ExecutePaymentResponseHandlerInterface
type ExecutePaymentResponseHandlerInterface interface {
	PublishPaymentResponse(executePaymentResponse commonMessagingModels.ExecutePaymentResponse) error
}

type ExecutePaymentResponseHandlerStruct struct {
}

var ExecutePaymentResponseHandler = func(ctx context.Context, message *kafkamessaging.Message) error {
	log.Info(ctx, "Received message in paymentplatform.internal.executepaymentresponse topic: '%s'", *message.Body)
	var paymentRepository repository.PaymentRepositoryInterface = repository.PaymentRepository{}
	var executePaymentResponseInterface ExecutePaymentResponseHandlerInterface = ExecutePaymentResponseHandlerStruct{}
	executePaymentResponse := commonMessagingModels.ExecutePaymentResponse{}
	if err := json.Unmarshal([]byte(*message.Body), &executePaymentResponse); err != nil {
		log.Error(context.Background(), err, "unable to unmarshal executePaymentResponse")
		return err
	}

	return ProcessExecutePaymentResponse(ctx, executePaymentResponse, paymentRepository, executePaymentResponseInterface)
}

func ProcessExecutePaymentResponse(
	ctx context.Context,
	executePaymentResponse commonMessagingModels.ExecutePaymentResponse,
	paymentRepository repository.PaymentRepositoryInterface,
	executePaymentResponseInterface ExecutePaymentResponseHandlerInterface) error {
	// Update payment status in database
	if err := paymentRepository.UpdatePaymentStatus(executePaymentResponse); err != nil {
		log.Error(context.Background(), err, "unable to update payment status")
		return err
	}

	// Get tenant info and send response to Kafka topic paymentPlatform.paymentResponses
	tenantId, tenantRequestId, metadata, err := paymentRepository.GetTenantInformationForPaymentId(executePaymentResponse.PaymentId)
	if err != nil {
		log.Error(context.Background(), err, "unable to create payment response")
		return err
	}

	// Remove PaymentId from response
	executePaymentResponse.PaymentId = 0
	executePaymentResponse.TenantId = tenantId
	executePaymentResponse.TenantRequestId = tenantRequestId
	executePaymentResponse.Metadata = metadata

	return executePaymentResponseInterface.PublishPaymentResponse(executePaymentResponse)
}

func (ExecutePaymentResponseHandlerStruct) PublishPaymentResponse(
	executePaymentResponse commonMessagingModels.ExecutePaymentResponse) error {

	executePaymentResponseJson, err := json.Marshal(executePaymentResponse)
	if err != nil {
		log.Error(context.Background(), err, "unable to marshal executePaymentResponse")
		return err
	}

	err = messaging.KafkaPublish(PaymentResponseTopic, string(executePaymentResponseJson))
	return err
}
