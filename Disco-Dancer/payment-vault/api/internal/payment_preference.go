package internal

import (
	"context"
	"fmt"

	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/repository"
	proto "geico.visualstudio.com/Billing/plutus/proto/paymentpreference"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var log = logging.GetLogger("payment-vault-api-internal")

func (s *Server) GetPaymentPreference(ctx context.Context, request *proto.PaymentPreferenceRequest) (*proto.PaymentPreferenceResponse, error) {
	userId := request.GetUserId()
	productIdentifier := request.GetProductIdentifier()
	transactionType := request.GetTransactionType()
	paymentRequestType := request.GetPaymentRequestType()

	transactionTypeEnum := enums.GetTransactionTypeEnum(transactionType)
	paymentRequestTypeEnum := enums.GetPaymentRequestTypeEnum(paymentRequestType)

	var paymentPreferenceRepository repository.PaymentPreferenceRepositoryInterface = repository.PaymentPreferenceRepository{}
	return getPaymentPreference(userId, productIdentifier, transactionTypeEnum, paymentRequestTypeEnum, paymentPreferenceRepository)

}

func getPaymentPreference(
	userId string,
	productIdentifier string,
	transactionTypeEnum enums.TransactionType,
	paymentRequestTypeEnum enums.PaymentRequestType,
	paymentPreferenceRepository repository.PaymentPreferenceRepositoryInterface) (*proto.PaymentPreferenceResponse, error) {

	if transactionTypeEnum == enums.NoneTransactionType {
		err := fmt.Errorf("invalid transaction type passed by client '%s'", transactionTypeEnum.String())
		log.Error(context.Background(), err, "")
		return &proto.PaymentPreferenceResponse{}, err
	}

	if paymentRequestTypeEnum == enums.NonePaymentRequestType {
		err := fmt.Errorf("invalid paymentrequest type passed by client '%s'", paymentRequestTypeEnum.String())
		log.Error(context.Background(), err, "")
		return &proto.PaymentPreferenceResponse{}, err
	}

	paymentPrefList, err := paymentPreferenceRepository.GetPaymentPreference(userId, productIdentifier, transactionTypeEnum, paymentRequestTypeEnum)

	if err != nil {
		log.Error(context.Background(), err, "error while getting payment preference")
		return &proto.PaymentPreferenceResponse{}, err
	}

	var PaymentPreferenceList []*proto.PaymentPreference

	for _, paymentPref := range paymentPrefList {
		paymentPreference := proto.PaymentPreference{
			PaymentMethodType:      enums.GetPaymentMethodTypeEnum(int(paymentPref.PaymentMethodType)).String(),
			AccountIdentifier:      paymentPref.AccountIdentifier,
			RoutingNumber:          paymentPref.RoutingNumber,
			PaymentExtendedData:    paymentPref.PaymentExtendedData,
			WalletStatus:           paymentPref.WalletStatus,
			PaymentMethodStatus:    enums.GetPaymentMethodStatusEnum(int(paymentPref.PaymentMethodStatus)).String(),
			AccountValidationDate:  timestamppb.New(paymentPref.AccountValidationDate),
			WalletAccess:           paymentPref.WalletAccess,
			AutoPayPreference:      paymentPref.AutoPayPreference,
			Split:                  paymentPref.Split,
			Last4AccountIdentifier: paymentPref.Last4AccountIdentifier,
		}
		PaymentPreferenceList = append(PaymentPreferenceList, &paymentPreference)
	}
	return &proto.PaymentPreferenceResponse{
		PaymentPreference: PaymentPreferenceList,
	}, nil
}
