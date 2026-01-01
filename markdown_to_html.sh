#! /bin/sh
#
# For local test of markdown generation, use a standalone Markdown-to-HTML processor.
#
# Plan for support of Goat diagrams embedded in Markdown locally,
# generated for production e.g. upload to a non-GH server
#
#   WITHOUT help from `go doc`:
#       1. Go filter preprocesses .md file.
#       2. Filter extracts and renders Goat diagrams to out-of-line** SVG files.
#       3. Filter replaces inline ASCII Goat diagram with relative link to SVG.
#
#     **  SVG content unlike PNG et al. is of course natively ASCII, but GFM apparently
#         nevertheless will not accept it inline, embedded.
#           c.f. https://github.github.com/gfm/#images
#
#   WITH help from `go doc`:
#      ???  X X  https://pkg.go.dev:
#                   1. grabs GH README.md, renders to HTML, discarding SVG.
#                   2. Runs `go doc` and appends its output.
#                                        ^^^^^^^ XX  So generating README.md from package
#                                                    would produce duplication!  
#      references:
#        no GitHub activity in many years from creator:
#              github.com/robertkrimen/godocdown/
#
#             XX  Fixed template -- no support for layering in new content e.g. diagrams.
#              github.com/davecheney/godoc2md
#        fork with later fixes, now *deleted* from GitHub UI, but `git pull` still works!
#              github.com/conusion/godoc2md

set -e
#set -x

# X  Deliberately "off" colors, as a reminder that this is no more than a proof rendering.
svg_color_light_scheme="#210"
svg_color_dark_scheme="#FFE"

# X Alternatives to 'marked':
#   -  https://github.com/yuin/goldmark
#          X X  https://github.com/abhinav/goldmark-toc
#                => Respin this very script as a Go CLI app, incorporating above libraries.
#   -  https://github.com/remarkjs/remark
#          XX  Coded in JS.
#          - https://github.com/remarkjs/remark-toc
#   -  https://github.com/gomarkdown/markdown
#          XX  No Table-Of-Contents support found.
#          - CLI: https://github.com/gomarkdown/mdtohtml

# See https://github.github.com/gfm/#introduction

#  The <!doctype html> below is as advised by:
#     https://developer.mozilla.org/en-US/docs/Web/HTML/Guides/Quirks_mode_and_standards_mode
#
# The @media query from SVG may be verified in Firefox by switching between Themes
#    "Light" and "Dark" in Firefox's "Add-ons Manager".
MARKED_PREAMBLE='<!doctype html>
<!-- CSS values specific to local "marked" CLI processor, not to Github -->
<style type="text/css">
  body {
    background-color: '$svg_color_dark_scheme';
    color: '$svg_color_light_scheme';
    font-family: sans;
  }
  a {
    fill: '$svg_color_light_scheme';   /* for cascade downward to <text> */
  }
  @media (prefers-color-scheme: dark) {
     body {
       background-color: '$svg_color_light_scheme';
       color: '$svg_color_dark_scheme';
     }
     a {
       fill: '$svg_color_dark_scheme';   /* for cascade downward to <text> */
     }
  }
  h4 {
    /* Tighten grouping of "func Foo()" et al. with following <code> */
    margin-block-end: 0.8em;
  }
  code {
    font-size: 11pt;
  }
</style>
'

echo "${MARKED_PREAMBLE}"
#
#marked --gfm "$@"

# Run './cmd/goldmark' instead of 'marked'
#    https://github.com/yuin/goldmark
MOD_FILE=$(go env GOMOD)
MODULE_ROOT=${MOD_FILE%/*}
go run "$MODULE_ROOT"/cmd/goldmark <"$1"
