#! /bin/sh
#
# Run all tests, and all pre-compilation build steps.
# Certain output files should be committed to the SCM archive.
#
# Recall that 'go get ...' performs compilation-proper for go, within
# the user's local environment.

set -e
set -x
usage () {
    printf "%s\n\n" "$*"
    printf "usage: %s [-g GitHub_Username] [-w]\n" ${0##*/}
    printf "\t%s\t%s\n" ""
    printf "\t%s\t%s\n" "$*"
    exit 1
}

# Define colors for SVG ~foreground~ seen on Github front page.
svg_color_dark_scheme="#EEF"
svg_color_light_scheme="#011"

GOMOD=$(go env GOMOD)
from_username=${GOMOD##*github.com/}
githubuser=${from_username%%/*}

TEST_ARGS=

while getopts hg:iw flag
do
    case $flag in
        h)  usage "";;
        g)  githubuser=${OPTARG};;  # Override guess based on GOMOD
	w)  TEST_ARGS=${TEST_ARGS}" -write";;
        \?) usage "unrecognized option flag";;
    esac
done

tmpl_expand () {
    go run ./cmd/tmpl-expand Root="." GithubUser=${githubuser} "$@"
}

#tmpl_expand <go.tmpl.mod >go.mod
#tmpl_expand <./cmd/goat/main.tmpl.go >./cmd/goat/main.go

# SVG examples/ regeneration.
#
# If the command fails due to expected changes in SVG output, rerun
# this script with "TEST_ARGS=-write" first on the command line.
# X  Results are used as "golden" standard for GitHub-side regression tests --
#    so arguments here must not conflict with those in "test.yml".
go test -run . -v \
   ${TEST_ARGS}

# build README.md
tmpl_expand <README.md.tmpl >README.md \
   $(bash -c 'echo ./examples/{trees,overlaps,line-decorations,line-ends,dot-grids,large-nodes,small-grids,big-grids,complicated}.{txt,svg}')

# '-d' writes ./awkvars.out
cat *.go |
    awk '
        /[<]goat[>]/ {p = 1; next}
        /[<][/]goat[>]/ {p = 0; next}
        p > 0 {print}' |
    tee goat.txt |
    go run ./cmd/goat \
	-svg-color-dark-scheme ${svg_color_dark_scheme} \
	-svg-color-light-scheme ${svg_color_light_scheme} \
	>goat.svg

# Render to HTML, for local inspection.
./markdown_to_html.sh README.md >README.html
./markdown_to_html.sh CHANGELOG.md >CHANGELOG.html
