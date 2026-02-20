//go:build nodedebug

package runtime

import (
	"fmt"
	"runtime"
)

func NodePath() (string, error) {
	return "", fmt.Errorf("node runtime not bundled for %s/%s (nodedebug build)", runtime.GOOS, runtime.GOARCH)
}

func NpmPath() (string, error) {
	return "", fmt.Errorf("npm not available (nodedebug build)")
}

func WebPath() (string, error) {
	return "", fmt.Errorf("web directory not bundled (nodedebug build)")
}
