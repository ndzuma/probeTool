//go:build !nodedebug

package runtime

import "embed"

//go:embed node-*
var nodeFS embed.FS

//go:embed web
var webFS embed.FS
