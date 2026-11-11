-- +goose Up
-- +goose StatementBegin
ALTER TABLE links
ADD COLUMN IF NOT EXISTS original_url text NOT NULL,
ADD COLUMN IF NOT EXISTS short_name text NOT NULL,
ADD COLUMN IF NOT EXISTS short_url text NOT NULL UNIQUE,
ADD COLUMN IF NOT EXISTS created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links 
DROP COLUMN IF EXISTS original_url,
DROP COLUMN IF EXISTS short_name,
DROP COLUMN IF EXISTS short_url,
DROP COLUMN IF EXISTS created_at;
-- +goose StatementEnd
