DROP TRIGGER IF EXISTS update_visitors_updated_at ON visitors;

DROP TRIGGER IF EXISTS update_visits_updated_at ON visits;

DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at_column ();

DROP TABLE IF EXISTS visitors;

DROP TABLE IF EXISTS visits;

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS residences;

DROP TABLE IF EXISTS blocks;

DROP TABLE IF EXISTS societies;

DROP TABLE IF EXISTS cities;

DROP TYPE IF EXISTS visitor_type;

DROP TYPE IF EXISTS user_role;
