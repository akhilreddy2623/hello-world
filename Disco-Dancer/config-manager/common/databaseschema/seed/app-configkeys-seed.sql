DO $$

BEGIN

    -- DataHub application config keys
    IF (SELECT COUNT(1) FROM public.configuration WHERE "Key" = 'PaymentPlatform.TaskManager.Url') = 0 THEN
            INSERT INTO public.configuration(
            "Description",
            "Key",
            "Value",
            "ScopeLevelId",
            "Environment",
            "Filters",
            "MetaData",
            "Tag",
            "AutoNotificationEnabled",
            "NotificationMode",
            "NotificationConfig",
            "IsActive",
            "EffectiveStartDate",
            "EffectiveEndDate",
            "CreatedDate",
            "CreatedBy",
            "UpdatedDate",
            "UpdatedBy")
                
            VALUES (
            'Task Manager URL',
            'PaymentPlatform.TaskManager.Url', 
            'https://bilpmt-tskapi-dv.geico.net/task',   
            2,
            'DV1',
            '{"Application": "data-hub","Tenant": "","Product": "","Vendor": ""}', 
            '{"Test": "test"}', 
            'Task Manager URL',
            true, 
            'Event', 
            '{"Brokers": "test","Topic": "test","Partition": 2,"WebHookUrl" :  ""}', 
            true, 
            '2022-01-01 00:00:00', 
            '2027-12-31 23:59:59', 
            CURRENT_TIMESTAMP, 
            'admin', 
            CURRENT_TIMESTAMP, 
            'admin'
        );
    END IF;

END $$