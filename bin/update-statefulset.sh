#! /bin/sh

COMMIT_MESSAGE=$1
./bin/update-deps.sh "$COMMIT_MESSAGE" \
  && kubectl rollout restart statefulset sotah-server-statefulset
