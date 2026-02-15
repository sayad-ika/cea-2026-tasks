ALTER TABLE bulk_opt_outs ADD COLUMN created_by_user_id UUID;
UPDATE bulk_opt_outs SET created_by_user_id = user_id;
ALTER TABLE bulk_opt_outs ALTER COLUMN created_by_user_id SET NOT NULL;
ALTER TABLE bulk_opt_outs ADD CONSTRAINT fk_bulk_opt_outs_created_by FOREIGN KEY (created_by_user_id) REFERENCES users(id);
ALTER TABLE bulk_opt_outs ADD COLUMN reason TEXT;
