//go:build nodedebug

package agent

import (
	"fmt"
)

var Files interface{} = nil

func Extract(destDir string) error {
	return fmt.Errorf("agent files not embedded in nodedebug build")
}
