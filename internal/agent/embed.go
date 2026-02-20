//go:build !nodedebug

package agent

import "embed"

//go:embed all:files
var Files embed.FS
