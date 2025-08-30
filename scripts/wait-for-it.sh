#!/bin/sh
# wait-for-it.sh script for Docker containers

set -e

host="$1"
shift
cmd="$@"

# Extract host from host:port format if needed
if echo "$host" | grep -q ":"; then
  host=$(echo "$host" | cut -d: -f1)
fi

until pg_isready -h "$host" -p 5432 -U app; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - executing command"
exec $cmd 