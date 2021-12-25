#! /bin/sh

COMMIT_MESSAGE=$1
./bin/update-deps.sh "$COMMIT_MESSAGE" \
  && gcloud builds submit --config ./cloudbuild-gcr.yaml
