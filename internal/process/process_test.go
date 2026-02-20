package process

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServerPIDFileOperations(t *testing.T) {
	testPID := 12345

	if err := WriteServerPID(testPID); err != nil {
		t.Errorf("WriteServerPID failed: %v", err)
	}

	pid, err := ReadServerPID()
	if err != nil {
		t.Errorf("ReadServerPID failed: %v", err)
	}

	if pid != testPID {
		t.Errorf("ReadServerPID = %d, want %d", pid, testPID)
	}

	if err := RemoveServerPID(); err != nil {
		t.Errorf("RemoveServerPID failed: %v", err)
	}

	if _, err := ReadServerPID(); err == nil {
		t.Error("ReadServerPID should fail after RemoveServerPID")
	}
}

func TestTrayPIDFileOperations(t *testing.T) {
	testPID := 54321

	if err := WriteTrayPID(testPID); err != nil {
		t.Errorf("WriteTrayPID failed: %v", err)
	}

	pid, err := ReadTrayPID()
	if err != nil {
		t.Errorf("ReadTrayPID failed: %v", err)
	}

	if pid != testPID {
		t.Errorf("ReadTrayPID = %d, want %d", pid, testPID)
	}

	if err := RemoveTrayPID(); err != nil {
		t.Errorf("RemoveTrayPID failed: %v", err)
	}

	if _, err := ReadTrayPID(); err == nil {
		t.Error("ReadTrayPID should fail after RemoveTrayPID")
	}
}

func TestServerPIDFileLocation(t *testing.T) {
	pidFile := ServerPIDFile()
	if pidFile == "" {
		t.Error("ServerPIDFile should not be empty")
	}

	cacheDir, _ := os.UserCacheDir()
	expectedDir := filepath.Join(cacheDir, "probeTool")
	if !filepath.IsAbs(pidFile) {
		t.Error("ServerPIDFile should return absolute path")
	}

	if filepath.Dir(pidFile) != expectedDir {
		t.Errorf("ServerPIDFile dir = %s, want %s", filepath.Dir(pidFile), expectedDir)
	}
}

func TestTrayPIDFileLocation(t *testing.T) {
	pidFile := TrayPIDFile()
	if pidFile == "" {
		t.Error("TrayPIDFile should not be empty")
	}

	cacheDir, _ := os.UserCacheDir()
	expectedDir := filepath.Join(cacheDir, "probeTool")
	if !filepath.IsAbs(pidFile) {
		t.Error("TrayPIDFile should return absolute path")
	}

	if filepath.Dir(pidFile) != expectedDir {
		t.Errorf("TrayPIDFile dir = %s, want %s", filepath.Dir(pidFile), expectedDir)
	}
}

func TestIsServerRunningNoPIDFile(t *testing.T) {
	RemoveServerPID()
	if IsServerRunning() {
		t.Error("IsServerRunning should return false when no PID file exists")
	}
}

func TestIsTrayRunningNoPIDFile(t *testing.T) {
	RemoveTrayPID()
	if IsTrayRunning() {
		t.Error("IsTrayRunning should return false when no PID file exists")
	}
}

func TestStopServerNoPIDFile(t *testing.T) {
	RemoveServerPID()
	err := StopServer()
	if err == nil {
		t.Error("StopServer should return error when no PID file exists")
	}
}

func TestStopTrayNoPIDFile(t *testing.T) {
	RemoveTrayPID()
	err := StopTray()
	if err == nil {
		t.Error("StopTray should return error when no PID file exists")
	}
}
