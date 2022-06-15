#!/bin/bash

set -e

working_dir="$PWD"

# Create a temporary director to generate protobuf Go files.
TMPDIR=$(mktemp -d)

PROJECTROOT="$PWD"
PROJECT="github.com/sun-asterisk-research/domain_exporter"

git clone --quiet "$PWD" "$TMPDIR/$PROJECT"

git diff > "$TMPDIR/patch"

cd $TMPDIR/$PROJECT
[ -s "$TMPDIR/patch" ] && git apply "$TMPDIR/patch"
go mod vendor
cd $TMPDIR

find $PROJECT -not -path "$PROJECT/vendor/*" -type d | \
  while read -r dir
  do
    # Ignore directories with no proto files.
    ls ${dir}/*.proto > /dev/null 2>&1 || continue
    protoc --proto_path="$PROJECT/vendor":. --go_out=. ${dir}/*.proto
  done

find $PROJECT -not -path "$PROJECT/vendor/*" -type f -name "*.pb.go" | \
  while read -r pbgofile
  do
    echo "$pbgofile"
    dst=${PROJECTROOT}/${pbgofile/$PROJECT\//}
    cp "$pbgofile" "$dst"
  done

cd $working_dir
rm -rf $TMPDIR
