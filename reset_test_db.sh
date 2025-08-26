#!/bin/bash

DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_NAME="greed"

TABLES="users,transactions,accounts,plaid_items,delegations,plaid_webhook_records,refresh_tokens,transaction_tags,transactions_to_tags,verification_records"

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "TRUNCATE $TABLES RESTART IDENTITY CASCADE;" || { echo "Reset failed"; exit 1; }

echo "Test database reset successfully."