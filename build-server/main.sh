#!/bin/bash

echo "Cloning github repo"
echo "$GIT_URL"

REPO_NAME=$(basename "$GIT_URL" .git)

git clone "$GIT_URL"

cd "$REPO_NAME"

PROJECT_ID="$PROJECT_ID"

railpack prepare . --plan-out railpack-plan.json --info-out railpack-info.json

docker buildx build \
  --build-arg BUILDKIT_SYNTAX="ghcr.io/railwayapp/railpack-frontend" \
  -t "deployment-${PROJECT_ID}" \
  -f ./railpack-plan.json \
  --load \
  .

