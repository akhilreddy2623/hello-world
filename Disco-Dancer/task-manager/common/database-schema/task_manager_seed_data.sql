INSERT INTO public.scheduled_task(
		"TaskId","TaskName", "Component", "DependsOn", "IsActive", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
	VALUES ('1','processpayments', 'administrator', '[]', 'true', now(), 'admin', now(), 'admin');

INSERT INTO public.scheduled_task(
		"TaskId","TaskName", "Component", "DependsOn", "IsActive", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
	VALUES ('2','settleachpayments', 'executor', '[1]', 'true', now(), 'admin', now(), 'admin');

INSERT INTO public.task_schedule(
	"ScheduleId", "TaskId", "ExecutionParameters", "Increment", "MaxRunTime", "Schedule", "NextRun", "Status", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
	VALUES (1, 1
	 ,'{"TaskId": 1,
	    "Year": "YYYY",
	    "Month": "MM",
	    "Day": "DD",
	    "ExecutionParameters": {
	        "PaymentMethodType": "ach",
	        "PaymentRequestType": "all",
	        "PaymentFrequency": "all" }}'
	,15 -- Increment in 15 minutes
    ,120 -- Max run time of 120 minutes
    ,'30 14 * * *' -- cron schedule to run the job at 2 PM every day
    , CURRENT_DATE + TIME '14:30' -- FIrst entry for today date and time 2:30 PM
	, 1, now(), 'admin', now(), 'admin');
