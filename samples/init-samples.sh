#!/bin/sh
# Initialize sample projects as bare git repos under /data/samples/
# This runs once at container startup if /data/samples/ doesn't exist yet.

SAMPLES_SRC="/app/samples"
SAMPLES_DST="/data/samples"

if [ -d "$SAMPLES_DST/node-app.git" ]; then
  echo "Samples already initialized, skipping."
  exit 0
fi

mkdir -p "$SAMPLES_DST"

for dir in "$SAMPLES_SRC"/*/; do
  name=$(basename "$dir")
  bare="$SAMPLES_DST/$name.git"

  echo "Initializing sample: $name"

  # Create a temp repo, commit files, then make a bare clone
  tmp="/tmp/sample-$name"
  rm -rf "$tmp"
  mkdir -p "$tmp"
  cp -r "$dir"* "$tmp/"
  cd "$tmp"
  git init -b main
  git config user.email "sample@mypaas.local"
  git config user.name "My PaaS"
  git add -A
  git commit -m "Initial sample: $name"

  # Create bare repo
  git clone --bare "$tmp" "$bare"
  rm -rf "$tmp"

  echo "  -> $bare"
done

echo "All samples initialized."
