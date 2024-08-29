CREATE DATABASE config_manager;

CREATE TABLE IF NOT EXISTS public."tenant"
(
    "TenantId" SERIAL, -- Datatype is int and it will be auto incremented
    "Name" character varying(100) NOT NULL UNIQUE,
    "Description" text,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_tenant" PRIMARY KEY ("TenantId")
);

CREATE TABLE IF NOT EXISTS public."application"
(
    "ApplicationId" SERIAL, -- Datatype is int and it will be auto incremented
    "Name" character varying(100) NOT NULL UNIQUE,
    "Description" text,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_application" PRIMARY KEY ("ApplicationId")
);

CREATE TABLE IF NOT EXISTS public."vendor"
(
    "VendorId" SERIAL, -- Datatype is int and it will be auto incremented
    "Name" character varying(100) NOT NULL UNIQUE,
    "Description" text,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_vendor" PRIMARY KEY ("VendorId")
);

CREATE TABLE IF NOT EXISTS public.product
(
    "ProductId" SERIAL, -- Datatype is int and it will be auto incremented
    "Name" character varying(100) NOT NULL UNIQUE,
    "Description" text,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_product" PRIMARY KEY ("ProductId")
);

CREATE TABLE IF NOT EXISTS public.configuration_scope_level
(
    "ScopeLevelId" SERIAL, -- Datatype is int and it will be auto incremented
    "Name" character varying(100) NOT NULL UNIQUE,
    "Description" text,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_configuration_scope_level" PRIMARY KEY ("ScopeLevelId")
);

CREATE TABLE IF NOT EXISTS public.configuration 
(
    "ConfigId" SERIAL, -- Datatype is int and it will be auto incremented
    "Description" text,   
    "Key"   character varying(50) NOT NULL,
    "Value" text ,
    "ScopeLevelId" int NOT NULL,
    "Environment"   character varying(50),
    "Filters" json,
    "MetaData" json,
    "Tag"   character varying(100) ,
    "AutoNotificationEnabled" boolean,
    "NotificationMode" character varying(32),
    "NotificationConfig" json,
    "IsActive" boolean NOT NULL,
    "EffectiveStartDate" timestamp without time zone,
    "EffectiveEndDate" timestamp without time zone,
    "CreatedDate"       timestamp without time zone,
    "CreatedBy"         character varying(50),
    "UpdatedDate"       timestamp without time zone,
    "UpdatedBy"         character varying(50),
    CONSTRAINT "pk_configuration" PRIMARY KEY ("ConfigId"),
    CONSTRAINT "fk_configuration_scope_level" FOREIGN KEY ("ScopeLevelId") REFERENCES public.configuration_scope_level ("ScopeLevelId")
);

CREATE TABLE IF NOT EXISTS public.configuration_history 
(
    
    "ConfigHistoryId" BIGSERIAL, -- Datatype is bigint and it will be auto incremented
    "HistoryTimestamp" timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "ConfigId" int, 
    "Description" text,   
    "Key"   character varying(50) ,
    "Value" text ,
    "ScopeLevelId" int NOT NULL,
    "Environment"   character varying(50) NOT NULL,
    "Filters" json,
    "MetaData" json,
    "Tag"   character varying(100) ,
    "AutoNotificationEnabled" boolean NOT NULL,
    "NotificationMode" character varying(32) NULL,
    "NotificationConfig" json,
    "IsActive" boolean NOT NULL,
    "EffectiveStartDate" timestamp without time zone NOT NULL,
    "EffectiveEndDate" timestamp without time zone NOT NULL,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,   
    CONSTRAINT "pk_configuration_history" PRIMARY KEY ("ConfigHistoryId") 
);

CREATE TABLE IF NOT EXISTS public.notification_status
(
    "NotificationStatusId" SERIAL, -- Datatype is int and it will be auto incremented
    "Name" character varying(100) NOT NULL UNIQUE,
    "Description" text,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_notification_status" PRIMARY KEY ("NotificationStatusId")
);

CREATE TABLE IF NOT EXISTS public.notification
(
    "NotificationId" SERIAL, -- Datatype is int and it will be auto incremented
    "NotificationDetail" json,
    "StatusId" int NOT NULL,
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT pk_notification PRIMARY KEY ("NotificationId")
);