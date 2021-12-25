#! /bin/sh

COMMIT_MESSAGE=$1
./bin/update-deps.sh "$COMMIT_MESSAGE" \
  && cd .. \
  && gcloud builds submit --config ./cloudbuild-gcr.yaml
