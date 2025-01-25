#!/usr/bin/env bash

DB_FILE="database.db"

deps=("uuidgen" "sqlite3")
for dep in ${deps[@]}; do
  # Check dep is installed
  if [[ ! $(type -P ${dep}) ]]; then
    echo "Missing dependency: ${dep}"
    exit 1
  fi
done

generate_random_string() {
  local LENGTH=$1
  tr -dc A-Za-z0-9 </dev/urandom | head -c $LENGTH
}

# Check database exist
if [ ! -f "$DB_FILE" ]; then
  echo "Database not found: $DB_FILE"
  exit 1
fi

# Number of users to create
NUM_USERS=10
if [ -n "$1" ]; then
  NUM_USERS=$1
fi

echo "Inserting $NUM_USERS dummy users into the database..."

# Insert dummy users
for ((i = 1; i <= NUM_USERS; i++)); do
  ID=$(uuidgen)
  TELEGRAM_ID=$((100000 + RANDOM % 900000))
  TELEGRAM_NAME="User_$(generate_random_string 5)"
  ADMIN=$((RANDOM % 2)) # Randomly set admin as 0 or 1

  sqlite3 "$DB_FILE" "INSERT INTO users (id, telegram_id, telegram_name, admin) VALUES ('$ID', $TELEGRAM_ID, '$TELEGRAM_NAME', $ADMIN);"
  echo "Inserted user $i: ID=$ID, Telegram ID=$TELEGRAM_ID, Name=$TELEGRAM_NAME, Admin=$ADMIN"
done

echo "Dummy users inserted successfully."
