#!/bin/bash

# Load environment variables from .env file (if it exists)
if [[ -f .env ]]; then
    export $(cat .env | xargs)
fi

# Check that MYSQL_DATABASE variable is set
if [[ -z "${DB_NAME}" ]]; then
    echo "DB_NAME environment variable not set."
    exit 1
fi

# Set default values for other environment variables if they are not already set
: ${DB_USER:=root}
: ${DB_PASSWORD:=password}
: ${DB_HOST:=localhost}
: ${DB_PORT:=3306}

echo "Dropping database ${DB_NAME}..."
mysql -u "${DB_USER}" -p"${DB_PASSWORD}" -h "${DB_HOST}" -P "${DB_PORT}" -e "DROP DATABASE IF EXISTS ${DB_NAME}; CREATE DATABASE ${MYSQL_DATABASE};"

echo "Running migrations..."
for f in ./data/sql/migrations/*.sql; do
    echo "Running migration $(basename "$f" .sql)..."
    mysql -u "${DB_USER}" -p"${DB_PASSWORD}" -h "${DB_HOST}" -P "${DB_PORT}" "${DB_NAME}" < "$f"
done

echo "Done."
