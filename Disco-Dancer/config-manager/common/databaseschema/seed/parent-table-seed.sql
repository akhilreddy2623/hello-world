DO $$

-- Inserting seed data for tenant table
BEGIN

    IF (SELECT COUNT(1) FROM public.configuration_scope_level WHERE "Name" = 'Global') = 0 THEN
        INSERT INTO public."configuration_scope_level" ("Name", "Description", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
        VALUES ('Global', 'Global', CURRENT_TIMESTAMP, 'Admin', CURRENT_TIMESTAMP, 'Admin');
    END IF;


    IF (SELECT COUNT(1) FROM public.configuration_scope_level WHERE "Name" = 'Application') = 0 THEN
        INSERT INTO public."configuration_scope_level" ("Name", "Description", "CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy")
        VALUES ('Application', 'Application', CURRENT_TIMESTAMP, 'Admin', CURRENT_TIMESTAMP, 'Admin');
    END IF;

END $$