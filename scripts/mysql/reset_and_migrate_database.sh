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

echo "Recreating database ${DB_NAME}..."
mysql -u "${DB_USER}" -p"${DB_PASSWORD}" -h "${DB_HOST}" -P "${DB_PORT}" -e "DROP DATABASE IF EXISTS ${DB_NAME}; CREATE DATABASE ${DB_NAME};"
echo "Database has been recreated."

echo "Running migrations and triggers..."
for f in ./data/sql/migrations/*.sql; do
    echo "Running migration $(basename "$f" .sql)..."
    mysql -u "${DB_USER}" -p"${DB_PASSWORD}" -h "${DB_HOST}" -P "${DB_PORT}" "${DB_NAME}" < "$f"
done
for f in ./data/sql/triggers/*.sql; do
    echo "Running migration $(basename "$f" .sql)..."
    mysql -u "${DB_USER}" -p"${DB_PASSWORD}" -h "${DB_HOST}" -P "${DB_PORT}" "${DB_NAME}" < "$f"
done
echo "Migrations executed successfully."

# Ask for user input
read -p "Do you want to populate with development seed? (y/n): " user_input

# Convert user input to lowercase
user_input=${user_input,,}

# Define valid responses
valid_responses=("y" "n")

# Check if the response is valid
while [[ ! " ${valid_responses[@]} " =~ " ${user_input} " ]]; do
    echo "Invalid input. Please enter either 'y' or 'n'."
    read -p "Do you want to populate with development seed? (y/n): " user_input
    user_input=${user_input,,}
done

if [[ $user_input == "y" ]]; then
    echo "Populating database with development seed..."
    for f in ./data/sql/dev/*.sql; do
        echo "Executing $(basename "$f" .sql)..."
        mysql -u "${DB_USER}" -p"${DB_PASSWORD}" -h "${DB_HOST}" -P "${DB_PORT}" "${DB_NAME}" < "$f"
    done
    echo "Database population executed successfully."
elif [[ $user_input == "n" ]]; then
    :
    # Do nothing
fi

echo "Done."
