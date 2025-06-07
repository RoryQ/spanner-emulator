# Google Cloud Spanner Emulator

This repository wraps the official 
[Google Cloud Spanner Emulator](https://github.com/GoogleCloudPlatform/cloud-spanner-emulator)
with a convenience function to create a spanner instance on startup.

## Usage
Set the `SPANNER_DATABASE_ID`, `SPANNER_INSTANCE_ID` and `SPANNER_PROJECT_ID` environment variables when running the image.
You can omit the database id if you just need an instance.
```sh
docker run --env SPANNER_DATABASE_ID=db \
  --env SPANNER_INSTANCE_ID=inst \
  --env SPANNER_PROJECT_ID=proj \
  -p 9010:9010 -p 9020:9020 \
  roryq/spanner-emulator:latest
```

---
Thanks to [jacksonjesse/pubsub-emulator](https://github.com/jacksonjesse/pubsub-emulator) for the idea.
