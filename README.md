# GoAT: Go ASCII Tool

## What **GoAT** Can Do For You

* From a chunky ASCII-art source drawing, render a pretty SVG output [file](#complicated),
  with the [goat](./cmd/goat) CLI command.

* Build publication-ready illustrated documentation, in Markdown, directly
  from comments in Go source code.
  The shell command [goatdocdown](./cmd/goatdocdown) reformats the output of ```go doc -all``` into Github-flavored Markdown, and any embedded goat-format drawings are processed into SVG for inclusion within the Markdown.

## You Will Also Need

#### Graphical- or Rectangle-oriented text editing capability
Both **vim** and **emacs** offer useful support.
In Emacs, see the built-in rectangle-editing commands, and ```picture-mode```.

#### A fixed-pitch font with 2:1 height:width ratio as presented by your editor and terminal emulator
Most fixed-pitch or "monospace" Unicode fonts maintain a 2:1 aspect ratio for
characters in the ASCII range,
and all GoAT drawing characters are ASCII.
However, certain Unicode graphical characters e.g. MIDDLE DOT may be useful, and
conform to the width of the ASCII range.

CJK characters on the other hand are typically wider than 2:1.
Non-standard width characters are not in general composable on the left-right axis within a plain-text
drawing, because the remainder of the line of text to their right is pushed out of alignment
with rows above and below.

## The GoAT Library [API](./API.md)

The core engine of ```goat``` is accessible as a Go library package, for inclusion in specialized
code of your own.
The [API documentation](./API.md) also showcases use of ```goatdocdown``` to generate illustrated Markdown directly from Go inline comments.

The code implements a subset, and some extensions, of the ASCII diagram generation function of the browser-side Javascript in [Markdeep](http://casual-effects.com/markdeep/).

## Project Tenets

1. Utility and ease of integration into existing projects are paramount.
2. Compatibility with MarkDeep desired, but not required.
3. TXT and SVG intelligibility are co-equal in priority.
4. Composability of TXT not to be sacrificed -- only width-8 characters allowed.
5. Per-platform support limited to a single widely-available fixed-pitch TXT font. 

## TODO

- Dashed lines signaled by `:` or `=`
- Bold lines signaled by ???

## Examples

Here are some snippets of 
GoAT-formatted UTF-8
and the SVG each can generate.
The SVG you see below was linked to by
inline Markdown image references
([howto](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#images),
[spec](https://github.github.com/gfm/#images)) from
GoAT's [README.md](README.md), then finally rendered to HTML ```<img>``` elements by Github's Markdown processor


### Trees
```

          .               .                .               .--- 1          .-- 1     / 1
         / \              |                |           .---+            .-+         +
        /   \         .---+---.         .--+--.        |   '--- 2      |   '-- 2   / \ 2
       +     +        |       |        |       |    ---+            ---+          +
      / \   / \     .-+-.   .-+-.     .+.     .+.      |   .--- 3      |   .-- 3   \ / 3
     /   \ /   \    |   |   |   |    |   |   |   |     '---+            '-+         +
     1   2 3   4    1   2   3   4    1   2   3   4         '--- 4          '-- 4     \ 4


```
![](./examples/trees.svg)

### Overlaps
```

           .-.           .-.           .-.           .-.           .-.           .-.
          |   |         |   |         |   |         |   |         |   |         |   |
       .---------.   .--+---+--.   .--+---+--.   .--|   |--.   .--+   +--.   .------|--.
      |           | |           | |   |   |   | |   |   |   | |           | |   |   |   |
       '---------'   '--+---+--'   '--+---+--'   '--|   |--'   '--+   +--'   '--|------'
          |   |         |   |         |   |         |   |         |   |         |   |
           '-'           '-'           '-'           '-'           '-'           '-'

```
![](./examples/overlaps.svg)

### Line Decorations
```
                ________                            o        *          *   .--------------.
   *---+--.    |        |     o   o      |         ^          \        /   |  .----------.  |
       |   |    '--*   -+-    |   |      v        /            \      /    | |  <------.  | |
       |    '----->       .---(---'  --->*<---   /      .+->*<--o----'     | |          | | |
   <--'  ^  ^             |   |                 |      | |  ^    \         |  '--------'  | |
          \/        *-----'   o     |<----->|   '-----'  |__|     v         '------------'  |
          /\                                                               *---------------'

```
![](./examples/line-decorations.svg)

### Line Ends
```
   o--o    *--o     /  /   *  o  o o o o   * * * *   o o o o   * * * *      o o o o   * * * *
   o--*    *--*    v  v   ^  ^   | | | |   | | | |    \ \ \ \   \ \ \ \    / / / /   / / / /
   o-->    *-->   *  o   /  /    o * v '   o * v '     o * v \   o * v \  o * v /   o * v /
   o---    *---
                                 ^ ^ ^ ^   . . . .   ^ ^ ^ ^   \ \ \ \      ^ ^ ^ ^   / / / /
   |  |   *  o  \  \   *  o      | | | |   | | | |    \ \ \ \   \ \ \ \    / / / /   / / / /
   v  v   ^  ^   v  v   ^  ^     o * v '   o * v '     o * v \   o * v \  o * v /   o * v /
   *  o   |  |    *  o   \  \

   <--o   <--*   <-->   <---      ---o   ---*   --->   ----      *<--   o<--   -->o   -->*


```
![](./examples/line-ends.svg)

### Dot Grids
```

  o o o o o  * * * * *  * * o o *    o o o      * * *      o o o     · * · · ·     · · ·
  o o o o o  * * * * *  o o o o *   o o o o    * * * *    * o * *    · * * · ·    · · · ·
  o o o o o  * * * * *  o * o o o  o o o o o  * * * * *  o o o o o   · o · · o   · · * * ·
  o o o o o  * * * * *  o * o o o   o o o o    * * * *    o * o o    · · · · o    · · * ·
  o o o o o  * * * * *  * * * * o    o o o      * * *      o * o     · · · · ·     · · *


```
Note that '·' above is not ASCII, but rather Unicode, the MIDDLE DOT character, encoded with UTF-8.
![](./examples/dot-grids.svg)

### Large Nodes
```

   .---.       .-.        .-.       .-.                                       .-.
   | A +----->| 1 +<---->| 2 |<----+ 4 +------------------.                  | 8 |
   '---'       '-'        '+'       '-'                    |                  '-'
                           |         ^                     |                   ^
                           v         |                     v                   |
                          .-.      .-+-.        .-.      .-+-.      .-.       .+.       .---.
                         | 3 +---->| B |<----->| 5 +---->| C +---->| 6 +---->| 7 |<---->| D |
                          '-'      '---'        '-'      '---'      '-'       '-'       '---'

```
![](./examples/large-nodes.svg)

### Small Grids
![](./examples/small-grids.svg)
```
       ___     ___      .---+---+---+---+---.     .---+---+---+---.  .---.   .---.
   ___/   \___/   \     |   |   |   |   |   |    / \ / \ / \ / \ /   |   +---+   |
  /   \___/   \___/     +---+---+---+---+---+   +---+---+---+---+    +---+   +---+
  \___/ b \___/   \     |   |   | b |   |   |    \ / \a/ \b/ \ / \   |   +---+   |
  / a \___/   \___/     +---+---+---+---+---+     +---+---+---+---+  +---+ b +---+
  \___/   \___/   \     |   | a |   |   |   |    / \ / \ / \ / \ /   | a +---+   |
      \___/   \___/     '---+---+---+---+---'   '---+---+---+---'    '---'   '---'


```

### Big Grids
```
    .----.        .----.
   /      \      /      \            .-----+-----+-----.
  +        +----+        +----.      |     |     |     |          .-----+-----+-----+-----+
   \      /      \      /      \     |     |     |     |         /     /     /     /     /
    +----+   B    +----+        +    +-----+-----+-----+        +-----+-----+-----+-----+
   /      \      /      \      /     |     |     |     |       /     /     /     /     /
  +   A    +----+        +----+      |     |  B  |     |      +-----+-----+-----+-----+
   \      /      \      /      \     +-----+-----+-----+     /     /  A  /  B  /     /
    '----+        +----+        +    |     |     |     |    +-----+-----+-----+-----+
          \      /      \      /     |  A  |     |     |   /     /     /     /     /
           '----'        '----'      '-----+-----+-----'  '-----+-----+-----+-----+


```
![](./examples/big-grids.svg)

### Complicated
```
+-------------------+                           ^                      .---.
|    A Box          |__.--.__    __.-->         |      .-.             |   |
|                   |        '--'               v     | * |<---        |   |
+-------------------+                                  '-'             |   |
                       Round                                       *---(-. |
  .-----------------.  .-------.    .----------.         .-------.     | | |
 |   Mixed Rounded  | |         |  / Diagonals  \        |   |   |     | | |
 | & Square Corners |  '--. .--'  /              \       |---+---|     '-)-'       .--------.
 '--+------------+-'  .--. |     '-------+--------'      |   |   |       |        / Search /
    |            |   |    | '---.        |               '-------'       |       '-+------'
    |<---------->|   |    |      |       v                Interior                 |     ^
    '           <---'      '----'   .-----------.              ---.     .---       v     |
 .------------------.  Diag line    | .-------. +---.              \   /           .     |
 |   if (a > b)     +---.      .--->| |       | |    | Curved line  \ /           / \    |
 |   obj->fcn()     |    \    /     | '-------' |<--'                +           /   \   |
 '------------------'     '--'      '--+--------'      .--. .--.     |  .-.     +Done?+-'
    .---+-----.                        |   ^           |\ | | /|  .--+ |   |     \   /
    |   |     | Join        \|/        |   | Curved    | \| |/ | |    \    |      \ /
    |   |     +---->  o    --o--        '-'  Vertical  '--' '--'  '--  '--'        +  .---.
 <--+---+-----'       |     /|\                                                    |  | 3 |
                      v                             not:line    'quotes'        .-'   '---'
  .-.             .---+--------.            /            A || B   *bold*       |        ^
 |   |           |   Not a dot  |      <---+---<--    A dash--is not a line    v        |
  '-'             '---------+--'          /           Nor/is this.            ---

```
![](./examples/complicated.svg))

### More examples [here](examples)

[@bep]: https://github.com/bep/goat/
[@dmacvicar]: https://github.com/dmacvicar/goat
[@sw46]: https://github.com/sw46/goat/
[SVG]: https://en.wikipedia.org/wiki/Scalable_Vector_Graphics
[markdeep.mini.js]: http://casual-effects.com/markdeep/
[v0.93.0]: https://github.com/gohugoio/hugo/releases/tag/v0.93.0
