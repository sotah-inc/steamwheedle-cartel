#! /bin/sh

cd extern/steamwheedle-cartel-server/app \
  && cd app/ && go get source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git && git status . \
  && go mod vendor && go mod tidy && git status .
