package integrationtest

const truncate_scheduled_task = `TRUNCATE TABLE public.scheduled_task RESTART IDENTITY CASCADE`

const insert_scheduled_task = `INSERT INTO public.scheduled_task("TaskId","TaskName", "Component", "DependsOn", "IsActive", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
								VALUES (1,'processpayments','administrator', '[]', true, NOW(),'AUVuser1' , NOW() ,'AUVuser1');`
