-- +goose Up
-- +goose StatementBegin
ALTER TABLE incidents ALTER COLUMN latitude TYPE DECIMAL(17, 15);
ALTER TABLE incidents ALTER COLUMN longitude TYPE DECIMAL(18, 15);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE incidents ALTER COLUMN latitude TYPE DECIMAL(10, 8);
ALTER TABLE incidents ALTER COLUMN longitude TYPE DECIMAL(11, 8);
-- +goose StatementEnd
