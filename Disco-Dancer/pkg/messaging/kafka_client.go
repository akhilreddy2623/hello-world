package messaging

import (
	"context"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/logging"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
)

var logger logging.Logger = logging.GetLogger("kafka_client")
var client *kafkamessaging.Client

func InitKafka(brokers []string, groupId string) error {

	connInfo := kafkamessaging.KafkaConnInfo{
		Brokers: brokers,
	}
	configHandler := commonFunctions.GetConfigHandler()

	certificatePassword := configHandler.GetString("PaymentPlatform.Kafka.PfxCertPassword", "")
	if certificatePassword != "" {
		connInfo.Mechanism = "mTLS"
		connInfo.EncodedCertificatePfx = configHandler.GetString("PaymentPlatform.Kafka.EncodedPfxCert", "")
		connInfo.CertificatePassword = certificatePassword
	}

	var err error
	client, err = kafkamessaging.NewClient(connInfo, groupId)
	if err != nil {
		logger.Error(context.Background(), err, "error creating kafka client")
		return err
	}
	return nil
}

func KafkaPublish(topicName string, message string) error {
	//TODO: Add a logic to retry if publish fails
	messgae := kafkamessaging.NewMessage(&message)
	err := client.Publish(context.Background(), topicName, messgae)
	if err != nil {
		logger.Error(context.Background(), err, "error publishing message with id '%s' and topic '%s'", messgae.MessageID, topicName)
		return err
	}

	logger.Info(context.Background(), "success publishing message with id '%s' and topic '%s'", messgae.MessageID, topicName)
	return nil
}

func KafkaSubscribe(topicName string, handler kafkamessaging.MessageHandlerFunc) error {
	err := client.Subscribe(context.Background(), topicName, kafkamessaging.MessageHandlerFunc(handler))
	if err != nil {
		logger.Error(context.Background(), err, "error subscribing for topic '%s'", topicName)
		return err
	}
	logger.Info(context.Background(), "success subscribing for topic '%s'", topicName)
	return nil
}
