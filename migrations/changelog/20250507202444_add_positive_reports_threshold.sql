-- +goose Up
-- +goose StatementBegin
BEGIN;
ALTER TABLE incident_types ADD COLUMN positive_reports_threshold INT;
UPDATE incident_types SET positive_reports_threshold = 5 WHERE id = 1;
UPDATE incident_types SET positive_reports_threshold = 10 WHERE id = 2;
UPDATE incident_types SET positive_reports_threshold = 10 WHERE id = 3;
UPDATE incident_types SET positive_reports_threshold = 5 WHERE id = 4;
UPDATE incident_types SET positive_reports_threshold = 5 WHERE id = 5;
COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE incident_types DROP COLUMN positive_reports_threshold;
-- +goose StatementEnd
