//go:build !nodedebug

package runtime

import "embed"

//go:embed all:node-*
var nodeFS embed.FS

//go:embed all:web
var webFS embed.FS
