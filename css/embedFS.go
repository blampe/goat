package css

import (
	"embed"
)

// From https://pkg.go.dev/embed
//   .. The patterns are interpreted relative to the package directory containing the source file ...

//go:embed */*.css
var FileSystem embed.FS
