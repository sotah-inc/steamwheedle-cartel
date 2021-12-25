#! /bin/sh

COMMIT_MESSAGE=$1
git add ./pkg \
  && git commit -m "$COMMIT_MESSAGE" \
  && git push origin HEAD \
  && ./update-server.sh
