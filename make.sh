#! /bin/sh

set -e
#set -x
usage () {
    printf "%s\n\n" "$*"
    printf "usage: %s [-g GitHub_Username] [-w]\n" ${0##*/}
    printf "\t%s\t%s\n" ""
    printf "\t%s\t%s\n" "$*"
    exit 1
}

build_variant=build

# Define colors for SVG ~foreground~ seen on Github front page.
svg_color_dark_scheme="#EDF"
svg_color_light_scheme="#014"

TEST_ARGS=

while getopts hg:iw flag
do
    case $flag in
        h)  usage "";;
        g)  githubuser=${OPTARG};;  # XXXX  At present only controls local debug output via `marked`
        i)  build_variant="install";;
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

(cd cmd/goat; go ${build_variant})

if [ ! "$githubuser" ]
then
    # Build README.html for local inspection.
    # See https://github.github.com/gfm/#introduction
    #
    # The @media query from SVG may be verified in Firefox by switching between Themes
    #    "Light" and "Dark" in Firefox's "Add-ons Manager".
    (
	printf '
<!DOCTYPE html>
<style>
	 html {
	      color: %s;
	      background-color: %s;
	  }
     @media (prefers-color-scheme: dark) {
	 html {
	      color: %s;
	      background-color: %s;
	 }
     }
     a[href] {
     	  color: currentColor;
     }
</style>
' ${svg_color_light_scheme} white \
  ${svg_color_dark_scheme} black

     marked -gfm README.md) >README.html
fi

# '-d' writes ./awkvars.out 
 <cmd/goat/main.go awk '
 		   /^[/][*]goat$/ || /^goat[*][/]$/ {p += 1; next}
		    p%2 == 1 {print}' |
     ./cmd/goat/goat -sls 'purple' -sds '#F82' >cmd/goat/main.svg
