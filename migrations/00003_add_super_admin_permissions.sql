-- +goose Up

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'super_admin';

-- +goose Down
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE name = 'super_admin');
