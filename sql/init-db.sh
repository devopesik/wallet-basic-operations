#!/bin/sh
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Создаём пользователя, если не существует
    DO
    \$do\$
    BEGIN
       IF NOT EXISTS (
          SELECT FROM pg_catalog.pg_roles
          WHERE  rolname = '$DB_USER') THEN

          CREATE ROLE $DB_USER LOGIN PASSWORD '$DB_PASSWORD';
       END IF;
    END
    \$do\$;

    -- Создаём БД, если не существует
    SELECT 'CREATE DATABASE $DB_NAME'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME');\gexec

    -- Даём пользователю права на БД
    GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;

    -- даём права на схему public внутри БД
    \connect $DB_NAME
    GRANT ALL ON SCHEMA public TO $DB_USER;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;
EOSQL