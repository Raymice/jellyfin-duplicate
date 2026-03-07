#!/bin/bash

# The script will now terminate immediately if any command exits with a non-zero status.
set -e

BIN_DIR="bin"
BIN_NAME="jellyfin-duplicate"
CURRENT_DIR=$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")
SOURCE_FILE="main.go"

# Move to the directory containing the script to be able to run the build commands
cd "$CURRENT_DIR"

# Clean bin directory
rm -r $BIN_DIR

# platform definitions: GOOS GOARCH extension(optional)
# each line will be read into variables inside the loop
platforms=(
    "windows amd64 .exe"
    "windows arm64 .exe"
    "darwin amd64"
    "darwin arm64"
    "linux amd64"
    "linux arm64"
)

for entry in "${platforms[@]}"; do
    # split into components, ext may be empty
    read -r GOOS GOARCH EXTENSION <<< "$entry"
    # ensure EXTENSION is not undefined
    EXTENSION=${EXTENSION:-}

    TARGET_FILE_NAME="$BIN_NAME-$GOOS-$GOARCH$EXTENSION"

    echo "🛠️ Building $TARGET_FILE_NAME ..."
    go build -o $BIN_DIR/$TARGET_FILE_NAME $SOURCE_FILE
    echo "✅ $TARGET_FILE_NAME built"
done

