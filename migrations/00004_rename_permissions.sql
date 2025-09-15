-- +goose Up
-- +goose StatementBegin

-- Renames permissions for clarity
UPDATE permissions SET name = 'owner:create.vehicle' WHERE name = 'vehicle.create';
UPDATE permissions SET name = 'admin:read.vehicle' WHERE name = 'vehicle.read';
UPDATE permissions SET name = 'driver:update.vehicle' WHERE name = 'vehicle.update';
UPDATE permissions SET name = 'owner:delete.vehicle' WHERE name = 'vehicle.delete';
UPDATE permissions SET name = 'owner:read.vehicle' WHERE name = 'driver.create';
UPDATE permissions SET name = 'admin:read.driver' WHERE name = 'driver.read';
UPDATE permissions SET name = 'driver:update.driver' WHERE name = 'driver.update';
UPDATE permissions SET name = 'driver:delete.driver' WHERE name = 'driver.delete';

-- Assign permissions to the 'vehicle_owner' role
INSERT INTO role_permissions (role_id, permission_id)
SELECT
	r.id, p.id
FROM
	roles r, permissions p
WHERE
	r.name = 'vehicle_owner' AND p.name IN (
		'owner:create.vehicle',
		'owner:delete.vehicle',
		'owner:read.vehicle'
	);

-- Assign permissions to the 'driver' role
INSERT INTO role_permissions (role_id, permission_id)
SELECT
	r.id, p.id
FROM
	roles r, permissions p
WHERE
	r.name = 'driver' AND p.name IN (
		'driver:update.vehicle',
		'driver:update.driver',
		'driver:delete.driver'
	);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Reverts the permission names back to their original state
UPDATE permissions SET name = 'vehicle.create' WHERE name = 'owner:create.vehicle';
UPDATE permissions SET name = 'vehicle,read' WHERE name = 'admin:read.vehicle';
UPDATE permissions SET name = 'vehicle.update' WHERE name = 'driver:update.vehicle';
UPDATE permissions SET name = 'vehicle.delete' WHERE name = 'owner:delete.vehicle';
UPDATE permissions SET name = 'driver.create' WHERE name = 'owner:read.vehicle';
UPDATE permissions SET name = 'driver.read' WHERE name = 'admin:read.driver';
UPDATE permissions SET name = 'driver.update' WHERE name = 'driver:update.driver';
UPDATE permissions SET name = 'driver.delete' WHERE name = 'driver:delete.driver';

DELETE FROM role_permissions
WHERE permission_id IN (
	SELECT id FROM permissions WHERE name IN (
		'owner:create.vehicle',
		'owner:delete.vehicle',
		'owner:read.vehicle'
	)
) AND role_id = (SELECT id FROM roles WHERE name = 'vehicle_owner');

DELETE FROM role_permissions
WHERE permission_id IN (
	SELECT id FROM permissions WHERE name IN (
		'driver:update.vehicle',
		'driver:update.driver',
		'driver:delete.driver'
	)
) AND role_id = (SELECT id FROM roles WHERE name = 'driver');

-- +goose StatementEnd
