#! /bin/sh
#
# On local dev box, "proof" README files for eventual appearance on GitHub.

set -e
set -x

PACKAGE_PATH=$(go list .)
UPSTREAM_OWNER=${PACKAGE_PATH#*/}
UPSTREAM_OWNER=${UPSTREAM_OWNER%/*}

GITHUB_REPOSITORY_OWNER=$USER
CURRENT_BRANCH_NAME=$(git-branch --show-current)
# If the current branch name contains the GitHub username of the owner of the upstream repo,
# assume the intention is to prepare and push a pull request.
# X  Used for "proofing" deployment to pkg.go.dev, from contributor's account.  Anything else?
if [ $(expr "$CURRENT_BRANCH_NAME" : ".*$UPSTREAM_OWNER") != 0 ]
then
    GITHUB_REPOSITORY_OWNER=$UPSTREAM_OWNER
fi

examples_DIR="./examples"   # may be empty, or non-existent
tmpl_expand () {
    go run "$TOPDIR"/cmd/tmpl-expand \
       examples_DIR="$examples_DIR" Github_Repository_Owner=${GITHUB_REPOSITORY_OWNER} \
       "$@"
}

# Generate README.md files for pushing to GitHub and,
# for ease of local proofing only, generate GFM-ish .html

#  for dir in ascii utf8
#  do
#      $(git-ls-files "$dir/_README/hello-world.txt")
#      $(git-ls-files "$dir/_README/hello-world.styled.txt")
#  done

TOPDIR=$(pwd)
prefix=README

# printf "Attempting to expand $PWD/%s\n" ${prefix}.md.tmpl
# tmpl_expand <${prefix}.md.tmpl >${prefix}.md
# 
# "$TOPDIR"/markdown_to_html.sh ${prefix}.md >_${prefix}.html

for dir in ascii utf8 .
do
    (
	cd $dir

	# X  git-ls-files will output paths to files in examples_DIR relative to $dir
	printf "Attempting to expand $PWD/%s\n" ${prefix}.md.tmpl
	tmpl_expand <${prefix}.md.tmpl >${prefix}.md $(git-ls-files "$examples_DIR*.txt")

	"$TOPDIR"/markdown_to_html.sh ${prefix}.md >_${prefix}.html
    )
done
