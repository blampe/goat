#! /bin/sh
#
# Run all tests, and all pre-compilation build steps.
# Certain output files should be committed to the SCM archive.
#
# Recall that an end-user eventually installing with 'go get ...' will
# trigger a compilation from source within the local environment,
# without help from this file.

set -e
set -x
usage () {
    set +x
    printf "%s\n\n" "$*"
    printf "usage: %s [-w]\n" "${0##*/}"
    printf "\t%s\n" ""
    printf "\t%s\n" "$*"
    exit 1
}

TEST_ARGS=
while getopts h:w flag
do
    case $flag in
        h)  usage "";;
	w)  TEST_ARGS=${TEST_ARGS}" -write";;
        \?) usage "unrecognized option flag";;
    esac
done

PACKAGE_PATH=$(go list .)
UPSTREAM_OWNER=${PACKAGE_PATH#*/}
UPSTREAM_OWNER=${UPSTREAM_OWNER%/*}

GITHUB_REPOSITORY_OWNER=$USER
CURRENT_BRANCH_NAME=$(git-branch --show-current)
# If the current branch name contains the GitHub username of the owner of the upstream repo,
# assume the intention is to prepare and push a pull request.
if [ $(expr "$CURRENT_BRANCH_NAME" : ".*$UPSTREAM_OWNER") != 0 ]
then
    GITHUB_REPOSITORY_OWNER=$UPSTREAM_OWNER
fi

tmpl_expand () {
    go run ./cmd/tmpl-expand Github_Repository_Owner=${GITHUB_REPOSITORY_OWNER} "$@"
}

#tmpl_expand <go.tmpl.mod >go.mod
#tmpl_expand <./cmd/goat/main.tmpl.go >./cmd/goat/main.go

# SVG examples/ regeneration.
#
# If the command fails due to expected changes in SVG output, rerun
# this script with "TEST_ARGS=-write".
# X  Results are used as "golden" standard for GitHub-side regression tests --
#    so arguments here must not conflict with those in "test.yml".
#   XX  How to share a single arg list shared between the two i.e. "DRY"?
go test -run . -v \
   ${TEST_ARGS}

# Build other SVG files; linked to by README.md but not used for regression test.
# Define colors for SVG ~foreground~ seen on Github front page.
svg_color_dark_scheme="#EEF"
svg_color_light_scheme="#011"
github_blue_color="#2F81F7"
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
#   Illustrate a workaround for lack of support in certain browsers e.g. Safari for
#   inheritance of CSS property 'color-scheme' from <img> elements downward to nested
#   <svg> elements.
#      - https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-color-scheme
go run ./cmd/goat <examples/trees.txt \
   -svg-color-dark-scheme ${github_blue_color} \
   -svg-color-light-scheme ${github_blue_color} \
   >trees.mid-blue.svg

# build README.md
#  X `tac` is a slightly sleazy way to get the .txt/.svg pairs listed in the
#     source/dest order as required by tmpl_expand.
tmpl_expand <README.md.tmpl >README.md $(git-ls-files examples | tac)

printf "\nTo install in local GOPATH:\n\t%s\n" "go install ./cmd/goat"
