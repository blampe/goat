#! /bin/sh

set -e
build_variant=build
if [ "$1" ]
then
    build_variant="$1"
    shift
fi

go ${build_variant}
go run . --markdown >README.md
marked -gfm README.md >README.html

