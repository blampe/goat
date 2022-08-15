# Changelog

## [Unreleased]

No changes yet.

## [0.5.0] - 2022-02-07

### Update (@blampe)

I hacked together GoAT a number of years ago while trying to embed some
diagrams in a Hugo project I was playing with. Through an odd twist of fate
GoAT eventually made its way into the upstream Hugo project, and if you're
using [v0.93.0] you can embed these diagrams natively. Neat!

My original implementation was certainly buggy and not on par with markdeep.
I'm grateful for the folks who've helped smooth out the rough edges, and I've
updated this project to reflect the good changes made in the Hugo fork,
including a long-overdue `go.mod`.

There's a lot I would like to do with this project that I will never get to, so
instead I recommend you look at these forks:
 * [@bep] is the fork currently used by Hugo, which I expect to be more active
   over time.
 * [@dmacvicar] has improved SVG/PNG/PDF rendering.
 * [@sw46] has implemented a really wonderful hand-drawn style worth checking
   out.

TODO
 - Dashed lines signaled by `:` or `=`
 - Bold lines signaled by ???

### Changed

* Merges changes made by @bep and @dmacvicar in their forks. This includes
  breaking changes to the CLI, and likely behavioral differences in parsing.


[Unreleased]: https://github.com/blampe/goat/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/blampe/goat/compare/ce4b402c34941d7ef3468ae70b84e9b05e7563f3...v0.5.0
