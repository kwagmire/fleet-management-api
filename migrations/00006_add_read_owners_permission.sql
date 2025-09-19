-- +goose Up
-- +goose StatementBegin
INSERT INTO permissions (name)
VALUES ('admin:read.owner') ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT
	r.id, p.id
FROM 
	roles r, permissions p
WHERE
	r.name = 'super_admin' AND p.name = 'admin:read.owner';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM role_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE name = 'admin:read.owner');

DELETE FROM permissions
WHERE name = 'admin:read.owner';
-- +goose StatementEnd
