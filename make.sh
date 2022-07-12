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
svg_color_dark_scheme="#EDF"
svg_color_light_scheme="#014"

TEST_ARGS=

while getopts hg:iw flag
do
    case $flag in
        h)  usage "";;
        g)  githubuser=${OPTARG};;  # XX  At present only controls local debug output via `marked`
	w)  TEST_ARGS=${TEST_ARGS}" -write";;
        \?) usage "unrecognized option flag";;
    esac
done

# SVG examples/ regeneration.
#
# If the command fails due to expected changes in SVG output, rerun
# this script with "TEST_ARGS=-write" first on the command line.
# XX  Better not to fail if the .txt source has changed.
go test -run . -v \
   -svg-color-dark-scheme ${svg_color_dark_scheme} \
   -svg-color-light-scheme ${svg_color_light_scheme} \
   ${TEST_ARGS}

# build README.md
go run ./cmd/tmpl-expand Root="." <README.md.tmpl >README.md \
   $(bash -c 'echo ./examples/{trees,overlaps,line-decorations,line-ends,dot-grids,large-nodes,small-grids,big-grids,complicated}.{txt,svg}')

go_to_markdown () {
    BASENAME=$1
    # input is a scan of the Go package in $CWD
    go doc -all |
         tee ${BASENAME}.go.doc.txt |
    # XX relative path assumes $CWD is the project root dir
    go run ./cmd/goatdoc -svgfilesprefix=${BASENAME} >${BASENAME}.md
}

# build API.md   XX  rename $(go list -f {{.Name}}).goatdoc.md  ?
go_to_markdown API

if [ ! "$githubuser" ]  # XX  Is this the right test
then
    # Render to HTML, for local inspection.
  ./markdown_to_html.sh README.md >README.html
  ./markdown_to_html.sh CHANGELOG.md >CHANGELOG.html
  ./markdown_to_html.sh API.md >API.html
fi
