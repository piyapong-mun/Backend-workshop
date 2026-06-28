#!/bin/sh

set -e # stop all execution if any command fails

# run migrations
echo "Running migrations..."
./migrate -database "postgresql://root:mypassword@postgres:5432/simple_bank?sslmode=disable" -path ./migration up

# start the app
echo "Starting app..."
# run all arguments that pass to this script
exec "$@"

