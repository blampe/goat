#! /bin/sh
set -e
usage () {
    printf "%s\n\n" "$*"
    printf "usage: %s [-g GitHub_Username]\n" ${0##*/}
    exit 1
}

while getopts g:h flag
do
    case $flag in
        g)  githubuser=${OPTARG};;
        h)  usage "";;
        \?) usage "unrecognized option flag";;
    esac
done

# SVG examples/ regeneration.
go test -run . -v -write

(cd cmd/goat; go install)

if [ "$githubuser" ]
then
    # Github home page README.md, specific to $githubuser:
    linktargetsub="s,{{\.Root}},https://cdn.rawgit.com/${githubuser}/goat/main,"
else
    # by default, build README.md for local inspection
    linktargetsub="s,{{.Root}},.,"
fi
sed "${linktargetsub}" README.md.tmpl >README.md

if [ ! "$githubuser" ]
then
    # build README.html for local inspection
    # See https://github.github.com/gfm/#introduction
    #
    # The @media query from SVG may be verified in Firefox by switching between Themes
    #    "Light" and "Dark" in Firefox's "Add-ons Manager".
    (echo '<!DOCTYPE html>'; marked -gfm README.md) >README.html
fi

