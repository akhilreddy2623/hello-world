package handlers

import (
	"context"
	"encoding/json"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/data-hub-common/repository"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
)

var PaymentEventBalancingHandler = func(ctx context.Context, message *kafkamessaging.Message) error {
	log.Info(ctx, "Received message in paymentplatform.paymentevents topic: '%s'", *message.Body)

	paymentEvent := commonMessagingModels.PaymentEvent{}
	if err := json.Unmarshal([]byte(*message.Body), &paymentEvent); err != nil {
		log.Error(context.Background(), err, "unable to unmarshal balancing payment event")
		return err
	}

	var paymentEventRepository repository.PaymentEventRepository = repository.PaymentEventRepository{}

	err := paymentEventRepository.AddPaymentEvent(&paymentEvent)

	if err != nil {
		log.Error(context.Background(), err, "error storing payment event into data-hub")
		return err
	}

	log.Info(ctx, "Payment event successfully recoreded into data-hub. PaymentId: '%d'", paymentEvent.PaymentId)
	return nil
}
