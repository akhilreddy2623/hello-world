--CREATE DATABASE payment_vault;


CREATE TABLE IF NOT EXISTS public."product_details"
(
    "ProductDetailId"   bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "UserId"            character varying(100) NOT NULL,
    "ProductType"       smallint,
    "ProductSubType"    smallint,
    "PaymentRequestType"  smallint,
    "ProductIdentifier" character varying(100) NOT NULL,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_product_details" PRIMARY KEY ("ProductDetailId")
);


CREATE TABLE IF NOT EXISTS public."payment_method"
(
    "PaymentMethodId"   bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "UserId"            character varying(100) NOT NULL,
    "CallerApp"         smallint NOT NULL,
    "PaymentMethodType" smallint NOT NULL,
    "AccountIdentifier" character varying(19) NOT NULL,
    "RoutingNumber"     character varying(9),
    "Last4AccountIdentifier" character varying(4) NOT NULL,
    "NickName"          character varying(50),
    "PaymentExtendedData" json NOT NULL,
    "WalletStatus"      boolean,
    "Status"            smallint,
    "AccountValidationDate" timestamp without time zone NOT NULL,
    "WalletAccess"      boolean,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_payment_method" PRIMARY KEY ("PaymentMethodId")   
);


CREATE TABLE IF NOT EXISTS public."payment_preference"
(
    "PreferenceId"      bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "ProductDetailId"   bigint NOT NULL,
    "PayIn"             json,
    "PayOut"            json,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    
    CONSTRAINT "pk_payment_preference" PRIMARY KEY ("PreferenceId"),
    CONSTRAINT "fk_product_details_payment_preference_productdetailid" FOREIGN KEY("ProductDetailId") REFERENCES public.product_details("ProductDetailId")
);


CREATE TABLE IF NOT EXISTS public."ach_validation_history"
(
    "AchValidationHistoryId"      bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "UserId"                      character varying(100) NOT NULL,
    "BankAccountNumber"           character varying(19) NOT NULL,
    "RoutingNumber"               character varying(9) NOT NULL,
    "Amount"                      money NOT NULL,
    "ResponseType"                character varying(1) NOT NULL,
    "ResponseCode"                character varying(10) NOT NULL,
    "ValidationStatus"            smallint NOT NULL,
    "RawResponse"                 json NOT NULL,
    "ProductIdentifier"           character varying(100),
    "CallerApp"                   smallint,
    "CreatedDate"                 timestamp without time zone NOT NULL,
    "CreatedBy"                   character varying(32) NOT NULL,
    "UpdatedDate"                 timestamp without time zone NOT NULL,
    "UpdatedBy"                   character varying(32) NOT NULL,
    
    CONSTRAINT "pk_ach_validation_history" PRIMARY KEY ("AchValidationHistoryId")
);


CREATE TABLE IF NOT EXISTS public."vault_config_mapper"
(
    "ConfigId"          bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "ProductType"       smallint NOT NULL,
    "PayBusinessOptions"    character varying(100),
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_vault_config_mapper" PRIMARY KEY ("ConfigId")
);