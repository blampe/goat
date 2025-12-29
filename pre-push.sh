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
    printf "\tTo debug source code for a single failing test case, follow this example:\n"
#    printf "\t\tdlv debug ./cmd/goat --redirect stdin:./utf8/examples/s.txt --redirect stdout:./utf8/examples/s.svg -- -utf8 -sds '#FFF'"
    printf "\t\tdlv debug ./cmd/goat -- -utf8 -sds '#FFF' -i ./utf8/examples/s.txt -o ./utf8/examples/s.svg\n"
    exit 1
}

TEST_ARGS=''
while getopts h:w flag
do
    case $flag in
        h)  usage "";;
	w)  TEST_ARGS=${TEST_ARGS}" -write";;
        \?) usage "unrecognized option flag";;
    esac
done

DELTA_DIR_NAME="_examples_new"    # XX  Presently hard-coded into Go source

for dir in ascii utf8
do
    (
	# SVG examples/ regeneration.
	#
	# If the command fails due to expected changes in SVG output, rerun
	# this script with "TEST_ARGS=-write".
	# Results are used as "golden" standard for GitHub-side regression tests,
	# so in the normal case let there be no arguments here other than control
	# of logging, conflict with test runs in "test.yml".

	if [ -d $dir/$DELTA_DIR_NAME ]
	then
	    printf "\n Pre-existing visual diffs found in  directory ./%s/%s/\n\n" $dir $DELTA_DIR_NAME
	fi

	go test ./$dir/. -run . -v ${TEST_ARGS}
    )
done

(
    # A reduced-quality fallback hack for lack of support in browsers other than Firefox for
    # inheritance of CSS property 'color-scheme' from <img> elements downward to nested
    # <svg> elements.
    #
    # See third, last, row of table here:
    #     https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-color-scheme#browser_compatibility
    github_blue_color="#2F81F7"
    go run ./cmd/goat \
       -i ./ascii/examples/trees.txt \
       -svg-color-dark-scheme=${github_blue_color} \
       -svg-color-light-scheme=${github_blue_color} \
       >./ascii/_README/trees.mid-blue.svg
)

go run ./cmd/goat \
   -i ./ascii/examples/complicated.txt \
   -o ./ascii/_README/complicated.motley.svg \
   embed:style/ascii.css embed:examples/css-legend.css

#  # X  Effective, but does not persist.  Required for `go run -cover ...`
#  go env GOCOVERDIR $PWD

# From:
#   https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-readmes?page=2#about-readmes
#
#       If you put your README file in your repository's hidden .github, root, or docs directory,
#       GitHub will recognize and automatically surface your README to repository visitors.
#        ...
#       When your README is viewed on GitHub, any content beyond 500 KiB will be truncated.
#
#   https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-readmes?page=2#relative-links-and-image-paths-in-markdown-files
#
#       GitHub will automatically transform your relative link or image path based on
#       whatever branch you're currently on, so that the link or path always works. The
#       path of the link will be relative to the current file. Links starting with /
#       will be relative to the repository root. You can use all relative link operands,
#       such as ./ and ../.

# XX  Incorporate these commands into the respective README.md.tmpl files,
#     with ./cmd/tmpl-expand "shelling out" to run `goat`?   

COLORS="embed:palette/bold.css"

# X  For ./README.md:
go run ./cmd/goat/       -io ./ascii/_README/hello-world.txt
go run ./cmd/goat/ -utf8 -io  ./utf8/_README/hello-world.txt
go run ./cmd/goat/ -utf8 -i ./utf8/_README/hello-world.txt -o ./utf8/_README/hello-world.motley.svg \
   ./css/examples/css-legend.css
go run ./cmd/goat/       -io ./ascii/_README/hello-world.styled.txt $COLORS embed:style/ascii.css
go run ./cmd/goat/ -utf8 -io  ./utf8/_README/hello-world.styled.txt $COLORS embed:style/utf8.css
#go run ./cmd/goat/       -io ./ascii/_README/hello-world.href.txt ./ascii/_README/hello-world.href.css
go run ./cmd/goat/ -utf8 -i ./utf8/_README/hello-world.styled.txt -o ./utf8/_README/hello-world.href.svg \
   $COLORS embed:style/utf8.css \
   ./utf8/_README/hello-world.href.css

# X  For ./ascii/README.md
go run ./cmd/goat/       -i ./ascii/examples/complicated.txt -o ./ascii/_README/complicated.svg \
   $COLORS embed:style/ascii.css \
   embed:examples/css-legend.css

# X  For ./utf8/README.md
go run ./cmd/goat/ -utf8 -i ./utf8/_README/dataflow.utf8.txt -o ./utf8/_README/dataflow.bold.svg \
   $COLORS embed:style/utf8.css \
   ./utf8/_README/dataflow.css
go run ./cmd/goat/ -utf8 -i ./utf8/_README/dataflow.utf8.txt -o ./utf8/_README/dataflow.earth.svg \
   embed:palette/earth.css embed:style/utf8.css \
   ./utf8/_README/dataflow.css

# Execution is not necessary for push to GitHub -- for "proofing" of README appearance, only
./proof-markdown.sh
