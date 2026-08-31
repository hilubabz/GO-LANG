-- +goose up 
ALTER TABLE users
ADD COLUMN is_chirpy_red BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose down
ALTER TABLE users
DROP COLUMN is_chirpy_red;