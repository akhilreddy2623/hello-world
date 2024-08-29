--CREATE DATABASE taskmanager;
CREATE TABLE IF NOT EXISTS public."scheduled_task"
(
    "TaskId"            smallint NOT NULL,
    "TaskName"          character varying(100) NOT NULL,
    "Component"         character varying(100) NOT NULL,
    "DependsOn"         JSONB NULL,
    "IsActive"          boolean NOT NULL,    
    "CreatedDate"       timestamp without time zone NOT NULL,
    "CreatedBy"         character varying(32) NOT NULL,
    "UpdatedDate"       timestamp without time zone NOT NULL,
    "UpdatedBy"         character varying(32) NOT NULL,
    CONSTRAINT "pk_scheduled_task" PRIMARY KEY ("TaskId")
);

CREATE TABLE IF NOT EXISTS public."task_execution"
(
    "TaskExecutionId"      	int NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),
    "TaskId"                smallint NOT NULL,
    "ExecutionDate"         date NOT NULL, 
    "ExecutionParameters"   JSONB NOT NULL,
    "StartTime"             timestamp without time zone NOT NULL,    
    "EndTime"               timestamp without time zone NULL,
    "Status"                smallint NOT NULL,
    "RecordsProcessed"      int NULL,
    "ErrorDetails"          JSONB NULL,
    "CreatedDate"           timestamp without time zone NOT NULL,
    "CreatedBy"             character varying(32) NOT NULL,
    "UpdatedDate"           timestamp without time zone NOT NULL,
    "UpdatedBy"             character varying(32) NOT NULL,
    CONSTRAINT "pk_task_execution" PRIMARY KEY ("TaskExecutionId"),
    CONSTRAINT "fk_task_execution_scheduled_task_TaskId" FOREIGN KEY("TaskId") REFERENCES public.scheduled_task("TaskId")
);


CREATE TABLE IF NOT EXISTS public."task_schedule"
(
    "ScheduleId"                smallint NOT NULL,
    "TaskId"                    smallint NOT NULL,
    "ExecutionParameters"       JSONB NOT NULL,
    "Increment"                 smallint NOT NULL,    
    "MaxRunTime"                smallint NOT NULL,
    "Schedule"                  character varying(16) NOT NULL,
    "NextRun"                   timestamp without time zone NOT NULL,
    "Status"                    smallint NOT NULL,  
    "CreatedDate"               timestamp without time zone NOT NULL,
    "CreatedBy"                 character varying(32) NOT NULL,
    "UpdatedDate"               timestamp without time zone NOT NULL,
    "UpdatedBy"                 character varying(32) NOT NULL,
    CONSTRAINT "pk_task_schedule" PRIMARY KEY ("ScheduleId"),
    CONSTRAINT "fk_task_schedule_scheduled_task_TaskId" FOREIGN KEY("TaskId") REFERENCES public.scheduled_task("TaskId")
);