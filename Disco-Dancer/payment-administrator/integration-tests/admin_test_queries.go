package integrationtest

const truncate_Payment = `TRUNCATE TABLE public.payment RESTART IDENTITY CASCADE`
const truncate_PaymentRequest = `TRUNCATE TABLE public.incoming_payment_request RESTART IDENTITY CASCADE`

const insert_PaymentRequest = `INSERT INTO incoming_payment_request("TenantRequestId", "TenantId", "UserId", "AccountId", "CallerApp", "ProductIdentifier", "Amount", "PaymentDate", "PaymentFrequency", "TransactionType", "PaymentRequestType", "Status", "PaymentExtractionSchedule", "Metadata", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
															VALUES (10, 10, 'IAA101', 'AUV456',2,'ALL' , 20.16, '2024-04-16',1, 2,2 , 2, '[{"date":"2024-06-10","amount":20.16}]', '{"ClaimNumber":123,"DocumentUrl":"anc.com"}', NOW(), 'payment-administrator', NOW(),'payment-administrator')`

const insert_Payment = `INSERT INTO payment("RequestId", "TenantRequestId", "UserId", "AccountId", "ProductIdentifier", "PaymentFrequency", "PaymentDate", "Amount", "Status", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
									VALUES (1, 10, 'IAA101', 'AUV456', 'ALL', 1, '2024-04-16',20.16 , 2, NOW(), 'payment-administrator', NOW(), 'payment-administrator')`

const getPaymentStatusRequest = `SELECT p."Status" AS PaymentStatus, i."Status" AS RequestStatus FROM public.payment AS p 
								 INNER JOIN public.incoming_payment_request AS i
								 ON p."RequestId" = i."RequestId"
								 WHERE "PaymentId" = $1`

const getPaymentValues = `SELECT "PaymentId","RequestId","TenantRequestId","UserId","ProductIdentifier","PaymentFrequency","Amount"::numeric,"Status" FROM public.payment WHERE "PaymentId" = $1`
const getPaymentRequestValues = `SELECT "RequestId","TenantId","TenantRequestId","UserId","ProductIdentifier","PaymentFrequency","TransactionType","PaymentRequestType","CallerApp","PaymentExtractionSchedule"::json->0->'amount' AS Amount,"Status" FROM public.incoming_payment_request WHERE "RequestId" = $1`

// Vault DB Statements
const truncate_ProductDetails = `TRUNCATE TABLE public.product_details RESTART IDENTITY CASCADE`
const truncate_PaymentMethod = `TRUNCATE TABLE public.payment_method RESTART IDENTITY CASCADE`
const truncate_PaymentPreference = `TRUNCATE TABLE public.payment_preference RESTART IDENTITY CASCADE`

const insert_ProductDetails = `INSERT INTO product_details("UserId", "ProductType", "ProductSubType", "PaymentRequestType", "ProductIdentifier", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
VALUES ('IAA101',1,1,2,'ALL','2024-04-16','payment-administrator','2024-04-16','payment-administrator')`

const insert_PaymentMethod = `INSERT INTO payment_method("UserId", "CallerApp", "PaymentMethodType", "AccountIdentifier", "RoutingNumber", "Last4AccountIdentifier", "PaymentExtendedData", "WalletStatus", "Status", "AccountValidationDate", "WalletAccess", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
VALUES ('IAA101',1,1,'accountIdentifier','123456789','1234','[]',true,1,'2024-04-16',true,'2024-04-16','payment-administrator','2024-04-16','payment-administrator')`

const insert_PaymentPreference = `INSERT INTO payment_preference("ProductDetailId", "PayIn", "PayOut", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
VALUES (1,'[{"PaymentMethodId":1,"Split":100,"AutoPayPreference":true}]','[{"PaymentMethodId":1,"Split":100,"AutoPayPreference":true}]','2024-04-16','payment-administrator','2024-04-16','payment-administrator')`

const insert_PaymentPreferenceWithNoSplit = `INSERT INTO payment_preference("ProductDetailId", "PayIn", "PayOut", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
VALUES (1,'[]','[]','2024-04-16','payment-administrator','2024-04-16','payment-administrator')`

const insert_PaymentPreferenceWithBadSplit = `INSERT INTO payment_preference("ProductDetailId", "PayIn", "PayOut", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
VALUES (1,'[{"PaymentMethodId":1,"Split":99,"AutoPayPreference":true}]','[{"PaymentMethodId":1,"Split":99,"AutoPayPreference":true}]','2024-04-16','payment-administrator','2024-04-16','payment-administrator')`

// Executor DB Statements
const truncate_ExecutionRequest = `TRUNCATE TABLE public.execution_request RESTART IDENTITY CASCADE`
const getExecutionRequest = `SELECT "ExecutionRequestId","PaymentId","Amount"::numeric::float,"Last4AccountIdentifier","PaymentRequestType","PaymentMethodType" FROM public.execution_request WHERE "PaymentId" = $1`
