CREATE TYPE user_role AS ENUM (
    'ADMIN',
    'SOCIETY_MANAGER',
    'SECURITY',
    'OWNER',
    'RESIDENT'
);

CREATE TYPE visitor_type AS ENUM (
    'DELIVERY',
    'MAINTENANCE',
    'GUEST',
    'CAB',
    'STAFF'
);

CREATE TABLE cities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
);

CREATE TABLE societies (
    id BIGSERIAL PRIMARY KEY,
    city_id INT NOT NULL REFERENCES cities(id),
    name VARCHAR(100) NOT NULL,
    address TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(city_id, name)
);

CREATE TABLE blocks (
    id BIGSERIAL PRIMARY KEY,
    society_id BIGINT NOT NULL REFERENCES societies(id),
    name VARCHAR(50) NOT NULL,
    UNIQUE(society_id, name)
);

CREATE TABLE residences (
    id BIGSERIAL PRIMARY KEY,
    number VARCHAR(20) NOT NULL,
    block_id BIGINT REFERENCES blocks(id),
    society_id BIGINT REFERENCES blocks(id),
    floor INTEGER NOT NULL,
    UNIQUE(block_id, number)
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    access_code CHAR(8) UNIQUE,
    role user_role NOT NULL,
    name VARCHAR(100),
    residence_id BIGINT REFERENCES residences(id),
    society_id BIGINT REFERENCES societies(id),
    is_active BOOLEAN NOT NULL DEFAULT false,
    device_id VARCHAR(255) UNIQUE,
    activated_by UUID REFERENCES users(id),
    activated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
);

CREATE TABLE visits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visitor_id UUID NOT NULL REFERENCES visitors(id),
    residence_id BIGINT REFERENCES residences(id),
    checked_in_by UUID NOT NULL REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    check_in_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    check_out_time TIMESTAMPTZ,
    purpose TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE visitors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    photo_url TEXT,
    type visitor_type NOT NULL,
    pre_approved_till DATE,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
);

CREATE INDEX idx_societies_city ON societies(city_id);
CREATE INDEX idx_blocks_society ON blocks(society_id);
CREATE INDEX idx_residences_block ON residences(block_id);
CREATE INDEX idx_users_residence ON users(residence_id) WHERE residence_id IS NOT NULL;
CREATE INDEX idx_users_society ON users(society_id) WHERE society_id IS NOT NULL;
CREATE INDEX idx_users_access_code ON users(access_code) WHERE access_code IS NOT NULL;
CREATE INDEX idx_visits_residence ON visits(residence_id);
CREATE INDEX idx_visits_check_in_time ON visits(check_in_time);
CREATE INDEX idx_visitors_phone ON visitors(phone);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_visits_updated_at
    BEFORE UPDATE ON visits
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_visitors_updated_at
    BEFORE UPDATE ON visitors
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
