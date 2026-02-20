package tray

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ndzuma/probeTool/internal/process"
)

func (m *Manager) waitForServer(timeoutSeconds int) error {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	checkURL := fmt.Sprintf("http://localhost:%s/api/health", process.ServerPort)

	for time.Now().Before(deadline) {
		resp, err := http.Get(checkURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("server failed to start within %d seconds", timeoutSeconds)
}
