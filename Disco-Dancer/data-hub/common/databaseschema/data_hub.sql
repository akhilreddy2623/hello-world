CREATE DATABASE data_hub;

CREATE TABLE IF NOT EXISTS public."payments"
(
    "PaymentId" bigint NOT NULL,
    "PaymentDate" timestamp without time zone NOT NULL,
    "Amount" money NOT NULL,
    "SettlementAmount" money,
    "PaymentRequestType" smallint NOT NULL,
    "PaymentMethodType" smallint NOT NULL,
	"LatestEvent" smallint NOT NULL,
    "LatestEventDateTime" timestamp without time zone NOT NULL,
    CONSTRAINT "pk_payment" PRIMARY KEY ("PaymentId")
);

CREATE TABLE IF NOT EXISTS public."payment_events"
(
    "EventId" bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    "PaymentId" bigint NOT NULL,
    "EventDateTime" timestamp without time zone NOT NULL,
    "EventType" smallint NOT NULL,
    CONSTRAINT "pk_events" PRIMARY KEY ("EventId")
);

CREATE INDEX "idx_paymentid" ON public."payment_events" ("PaymentId");

CREATE INDEX "idx_events_datetime" ON public."payment_events" USING BRIN ("EventDateTime");



