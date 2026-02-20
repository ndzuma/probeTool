package tray

import (
	"fmt"
	"net/http"
	"time"
)

func (m *Manager) waitForServer(timeoutSeconds int) error {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	checkURL := m.dashboardURL + "/api/health"

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
