#!/usr/bin/env bash

# Git root path
ROOT_PATH=$(git rev-parse --show-toplevel)
# Get Version from go.mod
VERSION=$(sed -nr 's/^go ([0-9]+\.[0-9]+).*$/\1/p' $ROOT_PATH/go.mod | head -n 1)

# Download go to .tools folder
mkdir -p $ROOT_PATH/.tools
rm -r $ROOT_PATH/.tools/go

# get release information
wget -q https://go.dev/doc/devel/release -O $ROOT_PATH/.tools/release
# Extract the latest minor release
LATEST_MINOR="$(sed -nr 's/^.*id=\"go1\.20\.([0-9]+)\".*$/\1/p' $ROOT_PATH/.tools/release | tail -1)"

echo "Downloading $VERSION.$LATEST_MINOR"
rm $ROOT_PATH/.tools/release

wget -q https://go.dev/dl/go$VERSION.$LATEST_MINOR.linux-amd64.tar.gz -O $ROOT_PATH/.tools/go$VERSION.$LATEST_MINOR.linux-amd64.tar.gz

# Extract
echo "Extracting to $ROOT_PATH/.tools/go"
tar -C $ROOT_PATH/.tools -mxf "$ROOT_PATH/.tools/go$VERSION.$LATEST_MINOR.linux-amd64.tar.gz"
rm "$ROOT_PATH/.tools/go$VERSION.$LATEST_MINOR.linux-amd64.tar.gz"

# Add to PATH
echo "run source $ROOT_PATH/hack/path.sh to add go to the PATH"
