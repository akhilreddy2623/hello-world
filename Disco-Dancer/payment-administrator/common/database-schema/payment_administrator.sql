-- CREATE DATABASE payment_administrator;

CREATE TABLE IF NOT EXISTS public."incoming_payment_request"
(
    "RequestId"         bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "TenantRequestId"   bigint NOT NULL,
    "TenantId"          bigint NOT NULL,
    "UserId"            character varying(100) NOT NULL,
    "AccountId"         character varying(100),
    "CallerApp"         smallint NOT NULL,
    "ProductIdentifier" character varying(100) NOT NULL,
    "Amount"            money,
    "PaymentDate"       timestamp without time zone,                
    "PaymentFrequency"  smallint NOT NULL,
    "TransactionType"   smallint NOT NULL,
    "PaymentRequestType"    smallint NOT NULL,
    "Status"            smallint NOT NULL,    
    "PaymentExtractionSchedule"   json NOT NULL,
    "Metadata"          json,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_incoming_payment_request" PRIMARY KEY ("RequestId")
);

CREATE INDEX IF NOT EXISTS "idx_incoming_payment_request_tenantid" ON public."incoming_payment_request" ("TenantId");

CREATE TABLE IF NOT EXISTS public."payment"
(
    "PaymentId"         bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "RequestId"         bigint NOT NULL,
    "TenantRequestId"   bigint NOT NULL,
    "UserId"            character varying(100) NOT NULL,
    "AccountId"         character varying(100),
    "ProductIdentifier" character varying(100) NOT NULL,
    "PaymentFrequency"  smallint NOT NULL ,
    "PaymentDate"       timestamp without time zone NOT NULL,
    "Amount"            money NOT NULL,
    "Status"            smallint NOT NULL,
    "IsSentToWorkday"   boolean  NOT NULL DEFAULT FALSE,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_payment" PRIMARY KEY ("PaymentId"),
    CONSTRAINT "fk_incoming_payment_request_requestid" FOREIGN KEY("RequestId") REFERENCES public.incoming_payment_request("RequestId")
);

CREATE INDEX IF NOT EXISTS "idx_payment_paymentdate" ON public."payment" ("PaymentDate");

ALTER TABLE IF EXISTS public."payment"
ADD COLUMN IF NOT EXISTS "IsSentToWorkday" boolean NOT NULL DEFAULT FALSE;




