#!/bin/bash

echo "Cloning github repo"
echo "$GIT_URL"

REPO_NAME=$(basename "$GIT_URL" .git)

git clone "$GIT_URL"

cd "$REPO_NAME"

npm install
npm run dev -- --host 0.0.0.0