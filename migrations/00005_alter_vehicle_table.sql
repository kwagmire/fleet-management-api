-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicles
DROP CONSTRAINT vehicles_owner_id_key;

DELETE FROM vehicles WHERE year < 2005;

ALTER TABLE vehicles
ADD CONSTRAINT check_vehicle_year CHECK (year >= 2005);

ALTER TABLE users
ADD CONSTRAINT check_user_role CHECK (role IN ('driver', 'vehicle_owner', 'super_admin'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles
ADD CONSTRAINT vehicles_owner_id_key UNIQUE (owner_id);

ALTER TABLE vehicles
DROP CONSTRAINT check_vehicle_year;

ALTER TABLE users
DROP CONSTRAINT check_user_role;
-- +goose StatementEnd
