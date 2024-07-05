-- cities
CREATE TABLE cities
(
    id      SERIAL PRIMARY KEY,
    name    VARCHAR NOT NULL,
    state   VARCHAR NOT NULL,
    country VARCHAR NOT NULL,
    lat     float,
    lng     float
);

-- societies
CREATE TABLE societies
(
    id                SERIAL PRIMARY KEY,
    name              VARCHAR UNIQUE NOT NULL,
    developer         VARCHAR        NOT NULL,
    max_residences    INT            NOT NULL,
    city_id           INT            NOT NULL,
    CONSTRAINT fk_city FOREIGN KEY (city_id) REFERENCES cities (id),
    access_revoked_at TIMESTAMP WITH TIME ZONE,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- blocks
CREATE TABLE blocks
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR NOT NULL,
    society_id INT     NOT NULL,
    CONSTRAINT fk_society FOREIGN KEY (society_id) REFERENCES societies (id)
);

-- residences
CREATE TABLE residences
(
    id         SERIAL PRIMARY KEY,
    number     INTEGER NOT NULL,
    society_id INT     NOT NULL,
    CONSTRAINT fk_society FOREIGN KEY (society_id) REFERENCES societies (id),
    block_id   INT,
    CONSTRAINT fk_block FOREIGN KEY (block_id) REFERENCES blocks (id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- generate resident unique id routine
-- CREATE OR REPLACE FUNCTION generate_resident_unique_id()
--     RETURNS TEXT AS
-- $$
-- DECLARE
--     chars  TEXT    := 'abcdefghijklmnopqrstuvwxyz';
--     result TEXT;
--     exists BOOLEAN := TRUE;
-- BEGIN
--     WHILE exists
--         LOOP
--             result := '';
--             FOR i IN 1..6
--                 LOOP
--                     result := result || substr(chars, floor(random() * 26 + 1)::int, 1);
--                 END LOOP;
--
--             -- Check if the generated string already exists
--             SELECT EXISTS(SELECT 1 FROM residents WHERE id = result) INTO exists;
--         END LOOP;
--     RETURN result;
-- END;
-- $$ LANGUAGE plpgsql;

-- users roles
CREATE TABLE user_roles
(
    level SERIAL PRIMARY KEY,
    role  VARCHAR(32)
);
INSERT INTO user_roles (level, role)
VALUES (1, 'resident'),
       (2, 'security'),
       (3, 'operator'),
       (4, 'manager'),
       (5, 'admin');

-- users
CREATE TABLE users
(
    id                SERIAL PRIMARY KEY,
    name              VARCHAR,
    email             VARCHAR UNIQUE,
    phone_number      VARCHAR(20) UNIQUE,
    -- for residence access level
    residence_id      INT,
    CONSTRAINT fk_residence FOREIGN KEY (residence_id) REFERENCES residences (id),
    -- for society access level
    society_id        INT,
    CONSTRAINT fk_society FOREIGN KEY (society_id) REFERENCES societies (id),
    role_level        INT NOT NULL             DEFAULT 1,
    CONSTRAINT fk_user_role FOREIGN KEY (role_level) REFERENCES user_roles (level),
    access_revoked_at TIMESTAMP WITH TIME ZONE,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- visitors
CREATE TABLE visitors
(
    id             SERIAL PRIMARY KEY,
    name           VARCHAR            NOT NULL,
    phone_number   VARCHAR(20) UNIQUE NOT NULL,
    photo          TEXT,
    purpose        VARCHAR            NOT NULL,
    is_preapproved BOOLEAN            NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMP WITH TIME ZONE    DEFAULT now(),
    updated_at     TIMESTAMP WITH TIME ZONE    DEFAULT now()
);

-- visits
CREATE TYPE residence_visit_status as ENUM ('accepted','rejected','pre-approved','security cleared');
CREATE TABLE residence_visits
(
    id           SERIAL PRIMARY KEY,
    residence_id INT NOT NULL,
    CONSTRAINT fk_residence FOREIGN KEY (residence_id) REFERENCES residences (id),
    visitor_id   INT NOT NULL,
    CONSTRAINT fk_visitor FOREIGN KEY (visitor_id) REFERENCES visitors (id),
    status       residence_visit_status,
    arrival_time TIMESTAMP WITH TIME ZONE DEFAULT now(),
    exit_time    TIMESTAMP WITH TIME ZONE
);

-- auth secrets
CREATE TABLE auth_secrets
(
    id           SERIAL PRIMARY KEY,
    email        VARCHAR,
    phone_number VARCHAR(20),
    secret       VARCHAR(32)              NOT NULL,
    is_used      BOOLEAN                  DEFAULT FALSE,
    expires_at   TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_phone_number ON auth_secrets (phone_number);
CREATE INDEX idx_email ON auth_secrets (email);
CREATE INDEX idx_expires_at ON auth_secrets (expires_at);
