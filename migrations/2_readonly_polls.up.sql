ALTER TABLE `spolls_polls` ADD COLUMN is_readonly BOOL NOT NULL DEFAULT FALSE AFTER create_date;