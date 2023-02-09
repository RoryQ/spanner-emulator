#!/usr/bin/env bash

set -e

docker build . -t verify-emulator

docker run --env SPANNER_DATABASE_ID=db \
  --env SPANNER_INSTANCE_ID=inst \
  --env SPANNER_PROJECT_ID=proj \
  --detach \
  --name verify \
  verify-emulator

sleep 10

docker logs verify &> verifylogs

cat verifylogs

docker stop verify > /dev/null
docker rm verify > /dev/null

echo verifying log output

grep "instance created" verifylogs
grep "Cloud Spanner emulator running." verifylogs
grep "REST server listening at 0.0.0.0:9020" verifylogs
grep "gRPC server listening at 0.0.0.0:9010" verifylogs
grep "database created" verifylogs

echo logs contain expected output