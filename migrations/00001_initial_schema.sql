-- +goose Up
-- This script sets up the database schema for the Fleet Management System MVP.
-- It creates all the necessary tables with their respective columns and constraints.
--

-- Enable the pgcrypto extension for password hashing if needed, though bcrypt is typically handled in the application layer.
-- CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 1. Roles Table: Defines the different user roles in the system.
CREATE TABLE roles (
	id SERIAL PRIMARY KEY,
	name VARCHAR(50) UNIQUE NOT NULL
);

-- 2. Permissions Table: Lists all possible actions a user can perform.
CREATE TABLE permissions (
	id SERIAL PRIMARY KEY,
	name VARCHAR(100) UNIQUE NOT NULL
);

-- 3. Role-Permissions Join Table: Links roles to their permissions.
CREATE TABLE role_permissions (
	role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
	permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
	PRIMARY KEY (role_id, permission_id)
);

-- 4. Users Table: Stores user authentication details.
CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	fullname VARCHAR(100) NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
	email VARCHAR(100) UNIQUE NOT NULL,
	role VARCHAR(50) REFERENCES roles(name) ON DELETE RESTRICT NOT NULL
);

CREATE INDEX ON users (role);

-- 5. Drivers Table: Stores driver-specific information.
CREATE TABLE drivers (
	user_id INTEGER UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	license_id VARCHAR(50) UNIQUE NOT NULL,
	assigned BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX ON drivers (user_id);

CREATE TABLE vehicle_owners (
	user_id INTEGER UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	fleet_size INTEGER NOT NULL CHECK (fleet_size >= 0)
);

CREATE INDEX ON vehicle_owners (user_id);

-- 6. Vehicles Table: Stores information about the fleet's vehicles.
CREATE TABLE vehicles (
	id SERIAL PRIMARY KEY,
	make VARCHAR(50) NOT NULL,
	model VARCHAR(50) NOT NULL,
	year INTEGER NOT NULL,
	license_plate VARCHAR(20) UNIQUE NOT NULL,
	status VARCHAR(20) NOT NULL CHECK (status IN ('available', 'in_use', 'maintenance', 'out_of_service')),
	driver_id INTEGER UNIQUE REFERENCES drivers(user_id) ON DELETE SET NULL, -- A vehicle can only have one driver at a time.
	owner_id INTEGER UNIQUE REFERENCES vehicle_owners(user_id) ON DELETE CASCADE
	--current_latitude DECIMAL(10, 8),
	--current_longitude DECIMAL(11, 8)
);

CREATE INDEX ON vehicles (driver_id);
CREATE INDEX ON vehicles (owner_id);
CREATE INDEX ON vehicles (license_plate);
CREATE INDEX ON vehicles (status);

--
-- Insert initial data to get started
--

INSERT INTO roles (name) VALUES
('super_admin'),
('vehicle_owner'),
('driver');

INSERT INTO permissions (name) VALUES
('vehicle.create'),
('vehicle.read'),
('vehicle.update'),
('vehicle.delete'),
('driver.create'),
('driver.read'),
('driver.update'),
('driver.delete');
--('location.read'),
--('location.update'),

-- Example: The admin role has all permissions.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'superadmin';


-- +goose Down

DROP TABLE IF EXISTS vehicles;
DROP TABLE IF EXISTS vehicle_owners;
DROP TABLE IF EXISTS drivers;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS role_permissions;

DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
