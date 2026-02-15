ALTER TABLE bulk_opt_outs DROP CONSTRAINT fk_bulk_opt_outs_created_by;
ALTER TABLE bulk_opt_outs DROP COLUMN created_by_user_id;
ALTER TABLE bulk_opt_outs DROP COLUMN reason;
