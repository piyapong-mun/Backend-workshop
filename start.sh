#!/bin/sh

set -e # stop all execution if any command fails

# run migrations
echo "Running migrations..."
source app.env
./migrate -database "$DB_SOURCE" -path ./migration up

# start the app
echo "Starting app..."
# run all arguments that pass to this script
exec "$@"

