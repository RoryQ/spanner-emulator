#!/usr/bin/env bash
set -euo pipefail

cleanup() {
    echo "Running cleanup..."
    rm -f verifylogs
    docker stop verify 2>/dev/null || true
}

# Handle normal exit and various signals
trap cleanup EXIT
trap cleanup SIGINT
trap cleanup SIGTERM

# Begin Verification
docker build . -t verify-emulator

docker stop verify 2>/dev/null || true

export SPANNER_EMULATOR_HOST=localhost:9010
export SPANNER_DATABASE_ID=db
export SPANNER_INSTANCE_ID=inst
export SPANNER_PROJECT_ID=proj

docker run --rm --env SPANNER_DATABASE_ID=$SPANNER_DATABASE_ID \
  --env SPANNER_INSTANCE_ID=$SPANNER_INSTANCE_ID \
  --env SPANNER_PROJECT_ID=$SPANNER_PROJECT_ID \
  --detach \
  --name verify \
  -p 9010:9010 \
  -p 9020:9020 \
  verify-emulator

# wait for emulator to start
MAX_SECONDS_WAIT=30
attempt=1
while [ $attempt -le $MAX_SECONDS_WAIT ]; do
    if docker logs verify 2>&1 | grep -q "Cloud Spanner emulator running."; then
        break
    fi
    sleep 1
    attempt=$((attempt + 1))
done

if [ $attempt -gt $MAX_SECONDS_WAIT ]; then
    echo "Timeout waiting for emulator to start"
    exit 1
fi

docker logs verify &> verifylogs

cat verifylogs

echo verifying log output

# Replace multiple greps with
expected_patterns=(
    "instance created"
    "Cloud Spanner emulator running."
    "REST server listening at 0.0.0.0:9020"
    "gRPC server listening at 0.0.0.0:9010"
    "database created"
)

for pattern in "${expected_patterns[@]}"; do
    if ! grep -q "$pattern" verifylogs; then
        echo "Error: Missing expected output: $pattern"
        exit 1
    fi
done

echo logs contain expected output

echo verifying database connection
go run ./tools/connect.go