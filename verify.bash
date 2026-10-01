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

expected_patterns=(
    "instance created"
    "Cloud Spanner emulator running."
    "REST server listening at 0.0.0.0:9020"
    "gRPC server listening at 0.0.0.0:9010"
    "database created"
)

# wait for emulator to start and initialize
MAX_SECONDS_WAIT=30
attempt=1
all_found=0
while [ $attempt -le $MAX_SECONDS_WAIT ]; do
    all_found=1
    for pattern in "${expected_patterns[@]}"; do
        if ! docker logs verify 2>&1 | grep -q "$pattern"; then
            all_found=0
            break
        fi
    done
    if [ $all_found -eq 1 ]; then
        break
    fi
    sleep 1
    attempt=$((attempt + 1))
done

docker logs verify &> verifylogs
cat verifylogs

if [ $all_found -ne 1 ]; then
    echo "Timeout waiting for emulator to start and initialize"
    for pattern in "${expected_patterns[@]}"; do
        if ! grep -q "$pattern" verifylogs; then
            echo "Error: Missing expected output: $pattern"
        fi
    done
    exit 1
fi

echo logs contain expected output

echo verifying database connection
go run ./tools/connect.go