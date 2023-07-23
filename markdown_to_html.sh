#! /bin/sh
#
#  For local test of markdown generation, use a standalone Markdown-to-HTML processor.

set -e
#set -x

svg_color_light_scheme="#320"
svg_color_dark_scheme="#FEE"

# XX An alternative to 'marked', for consideration:
#       https://github.com/gomarkdown/mdtohtml

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
