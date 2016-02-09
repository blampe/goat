# GoAT: Go ASCII Tool

This is a Go implementation of [markdeep.mini.js]'s ASCII diagram
generation.

## Example

This SVG:

![Complicated Eple](https://cdn.rawgit.com/blampe/goat/master/examples/complicated1.svg)

Was rendered from this input:

```
.-------------------.                           ^                      .---.
|    A Box          |__.--.__    __.-->         |        |  _ -        |   |
|                   |        '--'               v                      |   |
'-------------------'                                                  |   |
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
    |   |     +---->  o     -o-         '-'  Vertical  '--' '--'  '--  '--'        +  .---.
 <--+---+-----'       |     /|\                                                    |  | 3 |
                      v                             not:line    'quotes'        .-'   '---'
  .-.             .---+--------.            /            A || B   *bold*       |        ^
 |   |           |   Not a dot  |      <---+---<--    A dash--is not a line    v        |
  '-'             '---------+--'          /           Nor/is this.            ---
```

More examples are available [here](examples).

## Usage

```bash
$ go get github.com/blampe/goat
$ goat my-cool-diagram.txt > my-cool-diagram.svg
```

## TODO

* Dashed lines signaled by `:` or `=`.
* Bold lines signaled by ???.
* Draw half-steps (`_.-`) correctly.

[markdeep.mini.js]: http://casual-effects.com/markdeep/
