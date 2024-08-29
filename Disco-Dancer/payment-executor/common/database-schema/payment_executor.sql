-- CREATE DATABASE payment_executor;

CREATE TABLE IF NOT EXISTS public."execution_request"
(
    "ExecutionRequestId"  bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "TenantId"          bigint NOT NULL,
    "PaymentId"         bigint NOT NULL,
    "ConsolidatedId"    bigint NULL,
    "AccountIdentifier" character varying(19) NOT NULL,
    "RoutingNumber"     character varying(9) NULL,
    "Last4AccountIdentifier"    character varying(4) NOT NULL,
    "Amount"            money NOT NULL,
    "PaymentDate"       timestamp without time zone NOT NULL,                
    "PaymentFrequency"  smallint NOT NULL,
    "TransactionType"   smallint NOT NULL,
    "PaymentRequestType"    smallint NOT NULL,
    "PaymentMethodType"     smallint NOT NULL,
    "PaymentExtendedData"     jsonb NOT NULL,
    "RetryCount"        smallint NULL,
    "Status"            smallint NOT NULL,
    "SettlementIdentifier"  character varying(19) NULL,    
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_execution_request" PRIMARY KEY ("ExecutionRequestId")
);

ALTER TABLE IF EXISTS public."execution_request"
ADD COLUMN IF NOT EXISTS "Last4AccountIdentifier" character varying(4) NOT NULL;

CREATE TABLE IF NOT EXISTS public."consolidated_request"
(
    "ConsolidatedId"    bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "AccountIdentifier" character varying(19) NOT NULL,
    "RoutingNumber"     character varying(9) NULL,
    "Last4AccountIdentifier"    character varying(4) NOT NULL,
    "Amount"            money NOT NULL,
    "PaymentDate"       timestamp without time zone NOT NULL, 
    "PaymentRequestType"    smallint NOT NULL,
    "PaymentExtendedData"     jsonb NOT NULL,
    "RetryCount"        smallint NULL,
    "Status"            smallint NOT NULL,
    "SettlementIdentifier"  character varying(19) NULL,      
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_consolidated_request" PRIMARY KEY ("ConsolidatedId")
);

ALTER TABLE IF EXISTS public."consolidated_request"
ADD COLUMN IF NOT EXISTS "Last4AccountIdentifier" character varying(4) NOT NULL;

CREATE TABLE IF NOT EXISTS public."file_deduplication"
(
    "FileId"      	    int NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),
    "FileHash"          character varying(40) NOT NULL,
    "FilePath"          character varying(300) NOT NULL,
    "BusinessFileType"  character varying(100) NOT NULL,
    "Status"            character varying(32) NOT NULL,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_file_deduplication" PRIMARY KEY ("FileId")
);

CREATE TABLE IF NOT EXISTS public."record_deduplication"
(
    "RecordId"          bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "FileId"            int NOT NULL,
    "RecordHash"        character varying(40) NOT NULL,
    "Status"            character varying(32) NOT NULL,
    "RowNumber"         int NOT NULL,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_record_deduplication" PRIMARY KEY ("RecordId"),
    CONSTRAINT "fk_record_deduplication_file_deduplication_FileId" FOREIGN KEY("FileId") REFERENCES public.file_deduplication("FileId")
);

CREATE TABLE IF NOT EXISTS public."subro_payment"
(
    "SubroId"      	    bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "CreditDate"        character varying(8) NULL,
    "RemitterName"      character varying(300) NULL,
    "ClaimNumber"       character varying(16) NULL,
    "UniqueId"          character varying(500) NOT NULL,
    "AppliedAmount"     numeric(12,2) NULL,
    "CheckNumber"       character varying(50) NULL,
    "BatchNumber"       character varying(100) NULL,
    "CheckAmount"       numeric(12,2) NULL,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_subro_payment" PRIMARY KEY ("SubroId")
);


