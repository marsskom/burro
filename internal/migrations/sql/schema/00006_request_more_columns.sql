-- +goose Up
ALTER TABLE request ADD COLUMN scheme TEXT NOT NULL;
ALTER TABLE request ADD COLUMN path TEXT NOT NULL;
ALTER TABLE request ADD COLUMN proto TEXT NOT NULL;
ALTER TABLE request ADD COLUMN headers TEXT NOT NULL;
ALTER TABLE request ADD COLUMN cookies TEXT NOT NULL;
ALTER TABLE request ADD COLUMN query_params TEXT NOT NULL;
ALTER TABLE request ADD COLUMN content_length INTEGER NOT NULL;
ALTER TABLE request ADD COLUMN remote_addr TEXT NOT NULL;

ALTER TABLE request ADD COLUMN resp_status TEXT;
ALTER TABLE request ADD COLUMN resp_status_code INTEGER;
ALTER TABLE request ADD COLUMN resp_proto TEXT;
ALTER TABLE request ADD COLUMN resp_headers TEXT;
ALTER TABLE request ADD COLUMN resp_content_length INTEGER;

-- +goose Down
ALTER TABLE request DROP COLUMN scheme;
ALTER TABLE request DROP COLUMN path;
ALTER TABLE request DROP COLUMN proto;
ALTER TABLE request DROP COLUMN headers;
ALTER TABLE request DROP COLUMN cookies;
ALTER TABLE request DROP COLUMN query_params;
ALTER TABLE request DROP COLUMN content_length;
ALTER TABLE request DROP COLUMN remote_addr;

ALTER TABLE request DROP COLUMN resp_status;
ALTER TABLE request DROP COLUMN resp_status_code;
ALTER TABLE request DROP COLUMN resp_proto;
ALTER TABLE request DROP COLUMN resp_headers;
ALTER TABLE request DROP COLUMN resp_content_length;
