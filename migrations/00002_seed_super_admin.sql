-- +goose Up

INSERT INTO users (fullname, password_hash, email, role) VALUES
('Abdullah Olayiwola', '$2a$10$FLpd.ZMWd9cHBZxNaU3R8O.PFxGGDfC9Eq3Z6mg1MoF2ob9v6wrVG', 'kwagmire999@gmail.com', 'super_admin');

-- +goose Down
DELETE FROM users WHERE email = 'kwagmire999@gmail.com';`
