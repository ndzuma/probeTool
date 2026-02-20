package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	ServerPort = "37330"
	NextJSPort = "37331"
)

var (
	serverPIDFile string
	trayPIDFile   string
)

func init() {
	cacheDir, _ := os.UserCacheDir()
	appCacheDir := filepath.Join(cacheDir, "probeTool")
	os.MkdirAll(appCacheDir, 0755)
	serverPIDFile = filepath.Join(appCacheDir, "server.pid")
	trayPIDFile = filepath.Join(appCacheDir, "tray.pid")
}

func WriteServerPID(pid int) error {
	return os.WriteFile(serverPIDFile, []byte(strconv.Itoa(pid)), 0644)
}

func ReadServerPID() (int, error) {
	data, err := os.ReadFile(serverPIDFile)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func RemoveServerPID() error {
	return os.Remove(serverPIDFile)
}

func ServerPIDFile() string {
	return serverPIDFile
}

func WriteTrayPID(pid int) error {
	return os.WriteFile(trayPIDFile, []byte(strconv.Itoa(pid)), 0644)
}

func ReadTrayPID() (int, error) {
	data, err := os.ReadFile(trayPIDFile)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func RemoveTrayPID() error {
	return os.Remove(trayPIDFile)
}

func TrayPIDFile() string {
	return trayPIDFile
}

func IsServerRunning() bool {
	pid, err := ReadServerPID()
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func IsTrayRunning() bool {
	pid, err := ReadTrayPID()
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func StopServer() error {
	pid, err := ReadServerPID()
	if err != nil {
		return fmt.Errorf("server not running (no PID file)")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		RemoveServerPID()
		return fmt.Errorf("server process not found")
	}

	if err := process.Signal(os.Interrupt); err != nil {
		process.Kill()
	}

	RemoveServerPID()
	return nil
}

func StopTray() error {
	pid, err := ReadTrayPID()
	if err != nil {
		return fmt.Errorf("tray not running (no PID file)")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		RemoveTrayPID()
		return fmt.Errorf("tray process not found")
	}

	if err := process.Signal(os.Interrupt); err != nil {
		process.Kill()
	}

	RemoveTrayPID()
	return nil
}
