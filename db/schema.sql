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
    id             SERIAL PRIMARY KEY,
    name           VARCHAR UNIQUE NOT NULL,
    developer      VARCHAR        NOT NULL,
    max_residences INT            NOT NULL,
    city_id        INT            NOT NULL,
    CONSTRAINT fk_city FOREIGN KEY (city_id) REFERENCES cities (id),
    is_valid       BOOLEAN        NOT NULL  DEFAULT true,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- blocks
CREATE TABLE blocks
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR UNIQUE NOT NULL,
    society_id INT            NOT NULL,
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
CREATE OR REPLACE FUNCTION generate_resident_unique_id()
    RETURNS TEXT AS
$$
DECLARE
    chars  TEXT    := 'abcdefghijklmnopqrstuvwxyz';
    result TEXT;
    exists BOOLEAN := TRUE;
BEGIN
    WHILE exists
        LOOP
            result := '';
            FOR i IN 1..6
                LOOP
                    result := result || substr(chars, floor(random() * 26 + 1)::int, 1);
                END LOOP;

            -- Check if the generated string already exists
            SELECT EXISTS(SELECT 1 FROM residents WHERE id = result) INTO exists;
        END LOOP;
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- residents
CREATE TABLE residents
(
    id           VARCHAR PRIMARY KEY      DEFAULT generate_resident_unique_id(),
    residence_id INT     NOT NULL,
    CONSTRAINT fk_residence FOREIGN KEY (residence_id) REFERENCES residences (id),
    is_primary   BOOLEAN NOT NULL         DEFAULT false,
    is_valid     BOOLEAN NOT NULL         DEFAULT true,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- visitors
CREATE TABLE visitors
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name       VARCHAR NOT NULL,
    mobile     VARCHAR NOT NULL,
    photo      TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- visits
CREATE TYPE visit_status as ENUM ('accepted','rejected','pre-approved','security cleared');
CREATE TABLE visits
(
    id           SERIAL PRIMARY KEY,
    residence_id INT          NOT NULL,
    CONSTRAINT fk_residence FOREIGN KEY (residence_id) REFERENCES residences (id),
    status       VISIT_STATUS NOT NULL,
    arrival_time TIMESTAMP WITH TIME ZONE DEFAULT now(),
    exit_time    TIMESTAMP WITH TIME ZONE
);

-- agents
CREATE TYPE agent_role as ENUM ('admin','manager','security');
CREATE TABLE agents
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR NOT NULL,
    mobile     VARCHAR NOT NULL,
    society_id INT     NOT NULL,
    CONSTRAINT fk_society FOREIGN KEY (society_id) REFERENCES societies (id)
);
