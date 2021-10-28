#! /bin/sh

ORIGINAL_DIR=`pwd`
COMMIT_MESSAGE=$1
git add ./pkg \
  && git commit -m "$COMMIT_MESSAGE" \
  && git push origin HEAD \
  && cd extern/steamwheedle-cartel-server \
  && cd app/ && go get source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git && git status . \
  && go mod vendor && go mod tidy && git status . \
  && cd ../ && gcloud builds submit --config ./cloudbuild-gcr.yaml && git status . \
  && git add . && git commit -m 'Update to latest.' && git push origin HEAD \
  && cd $ORIGINAL_DIR && git add extern/ && git commit -m 'Misc.' && git push origin HEAD
