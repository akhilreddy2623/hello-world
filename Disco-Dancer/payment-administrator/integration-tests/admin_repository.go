package integrationtest

import (
	"context"

	dbmodels "geico.visualstudio.com/Billing/plutus/payment-administrator-common/models/db"
	executorDbModels "geico.visualstudio.com/Billing/plutus/payment-executor-common/models/db"

	"geico.visualstudio.com/Billing/plutus/database"
	"github.com/geico-private/pv-bil-frameworks/config"
	"github.com/jackc/pgx/v5"
)

func VerifyPaymentRequestStatus(paymentId int64) (int, int, error) {
	var PaymentStatus, RequestStatus int

	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getPaymentStatusRequest,
		paymentId)

	err := row.Scan(&PaymentStatus, &RequestStatus)

	if err != nil {
		return 0, 0, err
	}
	return PaymentStatus, RequestStatus, nil
}

func GetPaymentRequestFromDb(paymentId int64) (dbmodels.IncomingPaymentRequest, float32, int, error) {
	var amount float32
	var status int
	var paymentRequest dbmodels.IncomingPaymentRequest
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getPaymentRequestValues,
		paymentId)
	err := row.Scan(&paymentRequest.RequestId,
		&paymentRequest.TenantId,
		&paymentRequest.TenantRequestId,
		&paymentRequest.UserId,
		&paymentRequest.ProductIdentifier,
		&paymentRequest.PaymentFrequencyEnum,
		&paymentRequest.TransactionTypeEnum,
		&paymentRequest.PaymentRequestTypeEnum,
		&paymentRequest.CallerAppEnum,
		&amount,
		&status)
	return paymentRequest, amount, status, err
}

func GetPaymentFromDb(paymentId int64) (dbmodels.Payment, error) {
	var payment dbmodels.Payment
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getPaymentValues,
		paymentId)
	err := row.Scan(&payment.PaymentId,
		&payment.RequestId,
		&payment.TenantRequestId,
		&payment.UserId,
		&payment.ProductIdentifier,
		&payment.PaymentFrequencyEnum,
		&payment.Amount,
		&payment.Status)
	return payment, err
}

func GetExecutionRequestFromDB(paymentId int64) (executorDbModels.ExecutionRequest, error) {
	var executionRequest executorDbModels.ExecutionRequest
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getExecutionRequest,
		paymentId)
	err := row.Scan(
		&executionRequest.ExecutionRequestId,
		&executionRequest.PaymentId,
		&executionRequest.Amount,
		&executionRequest.Last4AccountIdentifier,
		&executionRequest.PaymentRequestType,
		&executionRequest.PaymentMethodType)
	return executionRequest, err
}

func SeedDataToPostgresTables(params ...string) error {

	tx, err := database.NewDbContext().Database.BeginTx(context.Background(), pgx.TxOptions{})

	if err != nil {
		return err
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback(context.Background())

	for _, v := range params {
		_, err = tx.Exec(context.Background(), v)
		if err != nil {
			return err
		}
	}
	// Commit the transaction.
	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func SeedDataToVault(params ...string) error {
	oldDbContext := database.GetDbContext()
	vaultTestConfig := config.NewConfigBuilder().
		AddJsonFile("../../payment-vault/api/config/appsettings.json").
		AddJsonFile("../../payment-vault/api/config/secrets.json").Build()
	database.Init(vaultTestConfig)
	newPool := database.NewPgxPool()
	database.SetDbContext(database.DbContext{Database: newPool})

	err := SeedDataToPostgresTables(params...)
	if err != nil {
		return err
	}
	// Closes the vault db connection and switches back to the admin db connection.
	database.GetDbContext().Database.Close()
	database.SetDbContext(*oldDbContext)
	return nil
}
