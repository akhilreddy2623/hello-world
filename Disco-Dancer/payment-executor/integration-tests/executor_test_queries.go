package integrationtest

import (
	"fmt"
	"time"
)

const layout = "2006-01-02"

var (
	today      = time.Now().Format(layout)
	twoDaysAgo = time.Now().AddDate(0, 0, -2).Format(layout)
)

func GetInsertConsolidatedRequestQueries(status int, records int) string {

	queries := ""
	for i := 1; i <= records; i++ {
		queries += fmt.Sprintf(`
        INSERT INTO public."consolidated_request" (
            "ConsolidatedId", "AccountIdentifier", "RoutingNumber", "Last4AccountIdentifier", "Amount",
            "PaymentDate", "PaymentRequestType", "PaymentExtendedData", "RetryCount",
            "Status", "SettlementIdentifier", "CreatedDate", "CreatedBy",
            "UpdatedDate", "UpdatedBy"
        )
        OVERRIDING SYSTEM VALUE
        VALUES (
            '%d', '|VJF4{', '021000021', '7891', 100.00,
            '%s', 2, '{"LastName": "InsuranceAutoAuctions", "FirstName": "AkhilReddy", "ACHAccountType": "checking"}', 0,
            %d, '24060634567891C', '%s', 'Akhil1',
            '%s', 'Akhil1'
        );`, 12345+i, twoDaysAgo, status, today, today)
	}

	return queries
}

func GetInsertExecutionRequestQueries(status int, records int) string {
	queries := ""
	for i := 1; i <= records; i++ {
		queries += fmt.Sprintf(`
        INSERT INTO public.execution_request (
            "ExecutionRequestId", "TenantId", "PaymentId", "ConsolidatedId",
            "AccountIdentifier", "RoutingNumber", "Last4AccountIdentifier",
            "Amount", "PaymentDate", "PaymentFrequency",
            "TransactionType", "PaymentRequestType", "PaymentMethodType",
            "PaymentExtendedData", "RetryCount", "Status",
            "SettlementIdentifier", "CreatedDate", "CreatedBy",
            "UpdatedDate", "UpdatedBy"
        )
        OVERRIDING SYSTEM VALUE
        VALUES (
            '%d', 1, %d, NULL,
            '|VJF4{', '021000021', '7891',
            100.00, '%s', 2,
            1, 2, 1,
            '{"LastName": "InsuranceAutoAuctions", "FirstName": "AkhilReddy", "ACHAccountType": "checking"}', 0, %d,
            '24060634567891C', '%s', 'Akhil1',
            '%s', 'Akhil1'
        );`, 12344+i, i, twoDaysAgo, status, today, today)
	}

	return queries
}

const get_settlement_identifier = `SELECT "SettlementIdentifier"
FROM public."consolidated_request"
WHERE "ConsolidatedId" = $1`

const getPaymentStatus = `SELECT "Status"
FROM public.execution_request
WHERE "PaymentId" = $1`

const getConsolidatedStatus = `SELECT "Status"
FROM public."consolidated_request"
WHERE "ConsolidatedId" = $1`

const truncate__consolidated_request = `TRUNCATE TABLE public.consolidated_request RESTART IDENTITY CASCADE`
const truncate_execution_request = `TRUNCATE TABLE public.execution_request RESTART IDENTITY CASCADE`
const truncate__file_deduplication = `TRUNCATE TABLE public.file_deduplication RESTART IDENTITY CASCADE`
const truncate_record_deduplication = `TRUNCATE TABLE public.record_deduplication RESTART IDENTITY CASCADE`
