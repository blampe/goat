#! /bin/sh
#
#  For local test of markdown generation, use a standalone Markdown-to-HTML processor.
#
# XX XX   Plan for support of Goat diagrams embedded in Markdown locally
#         generated for production e.g. upload to a non-GH server:
#       1. Go filter preprocesses .md file.
#       2. Filter extracts and renders Goat diagrams to out-of-line** SVG files.
#       3. Filter replaces inline ASCII Goat diagram with relative link to SVG.
#
#     **  SVG content unlike PNG et al. is of course natively ASCII, but GFM apparently
#         nevertheless will not accept it inline, embedded.
#           c.f. https://github.github.com/gfm/#images

set -e
#set -x

svg_color_light_scheme="#320"
svg_color_dark_scheme="#FEE"

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
#
# The @media query from SVG may be verified in Firefox by switching between Themes
#    "Light" and "Dark" in Firefox's "Add-ons Manager".
MARKED_PREAMBLE='<!-- CSS values specific to local "marked" CLI processor, not to Github -->
<style type="text/css">
  body {
    background-color: '$svg_color_dark_scheme';
    color: '$svg_color_light_scheme';
    font-family: sans;
  }
  a {
    color: '$svg_color_light_scheme';
  }
  @media (prefers-color-scheme: dark) {
     body {
       background-color: '$svg_color_light_scheme';
       color: '$svg_color_dark_scheme';
     }
     a {
       color: '$svg_color_dark_scheme';
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
marked --gfm "$@"
