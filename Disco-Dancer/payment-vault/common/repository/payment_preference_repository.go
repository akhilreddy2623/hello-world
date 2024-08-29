package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/models/db"
	"github.com/jackc/pgx/v5"
)

const (
	All                      = "ALL"
	paymentPreference_Insert = `
    INSERT INTO public.payment_preference ("ProductDetailId", "PayIn", "PayOut", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy") 
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING "PreferenceId"`
	productDetails_Insert = `INSERT INTO public.product_details ("UserId", "ProductType", "ProductSubType", "PaymentRequestType", "ProductIdentifier", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy") 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING "ProductDetailId"`
	paymentMethod_IsExist = `SELECT "PaymentMethodId", "Status" FROM public.payment_method WHERE "PaymentMethodId" IN (%s) AND "UserId"=$%d`
)

var repositoryLog = logging.GetLogger("payment-vault-repository")

//go:generate mockery --name PaymentPreferenceRepositoryInterface
type PaymentPreferenceRepositoryInterface interface {
	GetPaymentPreference(
		userId string,
		productIdentifier string,
		transactionTypeEnum enums.TransactionType,
		paymentRequestTypeEnum enums.PaymentRequestType) ([]PaymentPreference, error)
	StorePaymentPreference(model *db.StorePaymentPreference) error
}

type PaymentPreferenceRepository struct {
}

type PaymentPreference struct {
	PaymentMethodType      int16
	AccountIdentifier      string
	RoutingNumber          string
	PaymentExtendedData    string
	WalletStatus           bool
	PaymentMethodStatus    int16
	AccountValidationDate  time.Time
	WalletAccess           bool
	AutoPayPreference      bool
	Split                  int32
	Last4AccountIdentifier string
}

func (PaymentPreferenceRepository) GetPaymentPreference(
	userId string,
	productIdentifier string,
	transactionTypeEnum enums.TransactionType,
	paymentRequestTypeEnum enums.PaymentRequestType) ([]PaymentPreference, error) {

	if productIdentifier == "" || productIdentifier == "0" {
		productIdentifier = All
	}

	// Get the SQL for checking if payment preference exists in the database
	getPaymentPreferenceExistsDbQuery := getPaymentPreferenceExistsDbQuery(transactionTypeEnum)

	var exists bool
	error := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getPaymentPreferenceExistsDbQuery,
		userId,
		productIdentifier,
		paymentRequestTypeEnum.EnumIndex()).Scan(&exists)

	if error != nil {
		repositoryLog.Error(context.Background(), error, "error in fetching product details from payment vault")
		return nil, error
	}

	// Get the payment preferences for product identifier ALL if payment preferences does not exist
	if !exists {
		productIdentifier = All
	}

	// Get the SQL for fetching payment preferences based on transaction type
	getPaymentPreferenceDbQuery := getPaymentPreferenceDbQuery(transactionTypeEnum)

	rows, err := database.NewDbContext().Database.Query(
		context.Background(),
		getPaymentPreferenceDbQuery,
		userId,
		productIdentifier,
		paymentRequestTypeEnum.EnumIndex())

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in fetching payment prefrences from payment vault")
		return nil, err
	}

	var paymentPreferenceList []PaymentPreference

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			repositoryLog.Error(context.Background(), err, "error while iterating dataset while fetching payment prefrences from payment vault")
			return nil, err
		}

		paymentPreference := PaymentPreference{
			PaymentMethodType:      values[0].(int16),
			AccountIdentifier:      values[1].(string),
			RoutingNumber:          values[2].(string),
			PaymentExtendedData:    values[3].(string),
			WalletStatus:           values[4].(bool),
			PaymentMethodStatus:    values[5].(int16),
			AccountValidationDate:  values[6].(time.Time),
			WalletAccess:           values[7].(bool),
			AutoPayPreference:      values[8].(bool),
			Split:                  values[9].(int32),
			Last4AccountIdentifier: values[10].(string),
		}
		paymentPreferenceList = append(paymentPreferenceList, paymentPreference)
	}

	return paymentPreferenceList, nil
}

func (p PaymentPreferenceRepository) StorePaymentPreference(model *db.StorePaymentPreference) error {

	// Is Payment Method Exists
	if err := p.isPaymentMethodExists(p.getDistinctPaymentMethodIdList(model), model.ProductDetail.UserId); err != nil {
		return err
	}

	//  TODO : ProductDetails Exists and Payment Preference Exists , make update paymentpreference

	// Initiate Transaction. Store Product Details and Payment Preference should be commited together
	// OR rollback changes if any error occurs
	transaction, err := database.NewDbContext().Database.Begin(context.Background())
	defer transaction.Rollback(context.Background())

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in starting transaction for storing payment preference")
		return err
	}

	// Add Product Details
	if err := p.storeProductDetails(transaction, &model.ProductDetail); err != nil {
		return err
	}

	// Store Payment Preference
	if err := p.storePaymentPreferenceDetails(transaction, model); err != nil {
		return err
	}
	// If everything went well, commit the transaction
	transaction.Commit(context.Background())

	return nil

}

func (PaymentPreferenceRepository) storePaymentPreferenceDetails(transaction pgx.Tx, model *db.StorePaymentPreference) error {
	currentTimestamp := time.Now()

	//tx, err := database.NewDbContext().Database.Begin()

	dbError := transaction.QueryRow(context.Background(), paymentPreference_Insert,
		model.ProductDetail.ProductDetailId,
		model.PayIn,
		model.PayOut,
		currentTimestamp,
		model.ProductDetail.UserId,
		currentTimestamp,
		model.ProductDetail.UserId,
	).Scan(&model.PreferenceId)

	if dbError != nil && !(model.PreferenceId > 0) {
		repositoryLog.Error(context.Background(), dbError, "error in storing payment preference on payment_preference table")
		return errors.New(unhandledExceptionOccurred)
	}
	return nil
}

func getPaymentPreferenceExistsDbQuery(transactionTypeEnum enums.TransactionType) string {
	getPaymentPreferenceDbquery := getPaymentPreferenceDbQuery(transactionTypeEnum)
	var getPaymentPreferenceExists = `SELECT EXISTS(` + getPaymentPreferenceDbquery + `)`
	return getPaymentPreferenceExists
}

func getPaymentPreferenceDbQuery(transactionTypeEnum enums.TransactionType) string {
	var getPaymentPreference = ""

	if transactionTypeEnum == enums.PayIn {
		getPaymentPreference = `SELECT
			PM."PaymentMethodType",
	        PM."AccountIdentifier",
	        PM."RoutingNumber",
			PM."PaymentExtendedData"::varchar,
			PM."WalletStatus",
			PM."Status",
			PM."AccountValidationDate",
			PM."WalletAccess",
			(PS.PayIn->'AutoPayPreference')::text::bool as AutoPreference,
			(PS.PayIn->'Split')::text::int  as Split,
			PM."Last4AccountIdentifier"
		FROM
	  			public.payment_preference PP
		JOIN LATERAL
	  			json_array_elements(PP."PayIn") AS PS(PayIn) ON TRUE
	    INNER JOIN
	  			public.payment_method PM
	  				ON (PS.PayIn->'PaymentMethodId')::text::bigint = PM."PaymentMethodId"
	   INNER JOIN
	  			public.product_details PD
	  				ON PP."ProductDetailId" = PD."ProductDetailId"
		WHERE
	  			PD."UserId" = $1 and PD."ProductIdentifier"= $2 and PD."PaymentRequestType" = $3`
	} else if transactionTypeEnum == enums.PayOut {
		getPaymentPreference = `SELECT
			PM."PaymentMethodType",
	        PM."AccountIdentifier",
	        PM."RoutingNumber",
			PM."PaymentExtendedData"::varchar,
			PM."WalletStatus",
			PM."Status",
			PM."AccountValidationDate",
			PM."WalletAccess",
			(PS.PayOut->'AutoPayPreference')::text::bool as AutoPayPreference,
			(PS.PayOut->'Split')::text::int  as Split,
			PM."Last4AccountIdentifier"
		FROM
	  			public.payment_preference PP
		JOIN LATERAL
	  			json_array_elements(PP."PayOut") AS PS(PayOut) ON TRUE
		INNER JOIN
	  			public.payment_method PM
	  				ON (PS.PayOut->'PaymentMethodId')::text::bigint = PM."PaymentMethodId"
		INNER JOIN
	  			public.product_details PD
	  				ON PP."ProductDetailId" = PD."ProductDetailId"
		WHERE
	  			PD."UserId" = $1 and PD."ProductIdentifier"= $2 and PD."PaymentRequestType" = $3`
	}
	return getPaymentPreference
}

func (PaymentPreferenceRepository) storeProductDetails(transaction pgx.Tx, model *db.ProductDetail) error {

	currentTimestamp := time.Now()

	dbError := transaction.QueryRow(context.Background(), productDetails_Insert,
		model.UserId,
		model.ProductType.EnumIndex(),
		model.ProductSubType.EnumIndex(),
		model.PaymentRequestType.EnumIndex(),
		model.ProductIdentifier,
		currentTimestamp,
		model.UserId,
		currentTimestamp,
		model.UserId,
	).Scan(&model.ProductDetailId)

	if dbError != nil {
		repositoryLog.Error(context.Background(), dbError, "error in storing  product details on payment_vault table")
		return errors.New(unhandledExceptionOccurred)
	}
	return nil
}

func (PaymentPreferenceRepository) isPaymentMethodExists(paymentMethodIds []int64, userId string) error {

	// Prepare PaymentMethodId list for query
	placeholders := make([]string, len(paymentMethodIds))
	for i := range paymentMethodIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(paymentMethod_IsExist, strings.Join(placeholders, ","), len(paymentMethodIds)+1)

	args := make([]interface{}, len(paymentMethodIds)+1)
	for i, id := range paymentMethodIds {
		args[i] = id
	}
	args[len(paymentMethodIds)] = userId

	rows, err := database.NewDbContext().Database.Query(context.Background(), query, args...)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error while fetching payment method details from payment vault")
		return err
	}
	defer rows.Close()

	// Ensure Payment Method is Exists and Active
	foundIds := make(map[int64]bool)
	for rows.Next() {
		var id int64
		var status int32
		if err := rows.Scan(&id, &status); err != nil {
			repositoryLog.Error(context.Background(), err, "error while scanning row")
			return err
		}
		if status != int32(enums.Active) {
			return fmt.Errorf("PaymentMethodId %d is not active", id)
		}
		foundIds[id] = true
	}

	if err := rows.Err(); err != nil {
		repositoryLog.Error(context.Background(), err, "error after iterating over rows")
		return err
	}

	for _, id := range paymentMethodIds {
		if !foundIds[id] {
			return fmt.Errorf("PaymentMethodId %d not found", id)
		}
	}

	return nil

}

func (PaymentPreferenceRepository) getDistinctPaymentMethodIdList(model *db.StorePaymentPreference) []int64 {
	distinctIds := make(map[int64]bool)

	for _, m := range model.PayIn {
		distinctIds[m.PaymentMethodId] = true
	}

	for _, m := range model.PayOut {
		distinctIds[m.PaymentMethodId] = true
	}

	var distinctPaymentMethodIds []int64
	for id := range distinctIds {
		distinctPaymentMethodIds = append(distinctPaymentMethodIds, id)
	}
	return distinctPaymentMethodIds
}
