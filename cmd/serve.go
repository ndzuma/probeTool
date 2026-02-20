package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ndzuma/probeTool/internal/db"
	"github.com/ndzuma/probeTool/internal/process"
	"github.com/ndzuma/probeTool/internal/runtime"
	"github.com/ndzuma/probeTool/internal/server"
	"github.com/spf13/cobra"
)

var quietMode bool
var daemonMode bool

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the probe dashboard server",
	Long:  `Starts both the API server and Next.js frontend.`,
	Run: func(cmd *cobra.Command, args []string) {
		runServe()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().BoolVar(&quietMode, "quiet", false, "Run in background (daemon mode)")
	serveCmd.Flags().BoolVar(&daemonMode, "daemon", false, "Run as daemon (internal use)")
}

func runServe() {
	if daemonMode {
		runDaemon()
		return
	}

	if quietMode {
		startAsDaemon()
		return
	}

	runForeground()
}

func startAsDaemon() {
	if process.IsServerRunning() {
		fmt.Println("Server is already running")
		return
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command(execPath, "serve", "--daemon")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server started in background")
	fmt.Println("Use 'probe stop' to stop the server")
}

func runDaemon() {
	runForeground()
}

func runForeground() {
	if process.IsServerRunning() {
		if !daemonMode {
			fmt.Println("Server is already running")
		}
		return
	}

	if err := process.WriteServerPID(os.Getpid()); err != nil {
		if !daemonMode {
			fmt.Printf("Warning: could not write PID file: %v\n", err)
		}
	}

	database, err := db.InitDB(db.DBPath())
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		process.RemoveServerPID()
		os.Exit(1)
	}
	defer database.Close()

	nodePath, err := runtime.NodePath()
	if err != nil {
		fmt.Printf("Error loading Node.js runtime: %v\n", err)
		process.RemoveServerPID()
		os.Exit(1)
	}

	webPath, err := runtime.WebPath()
	if err != nil {
		fmt.Printf("Error loading web assets: %v\n", err)
		process.RemoveServerPID()
		os.Exit(1)
	}

	if err := ensureNextJSReady(nodePath, webPath); err != nil {
		fmt.Printf("Error setting up Next.js: %v\n", err)
		process.RemoveServerPID()
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nextJSCmd := startNextJS(ctx, nodePath, webPath)
	if nextJSCmd == nil {
		fmt.Println("Failed to start Next.js server")
		process.RemoveServerPID()
		os.Exit(1)
	}
	defer func() {
		if nextJSCmd.Process != nil {
			nextJSCmd.Process.Kill()
		}
	}()

	if !waitForNextJS(process.NextJSPort, 30*time.Second) {
		fmt.Println("Next.js server failed to start")
		process.RemoveServerPID()
		os.Exit(1)
	}

	mux := createUnifiedServer(database)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		if !daemonMode {
			fmt.Println("\nShutting down gracefully...")
		}
		cancel()
		if nextJSCmd != nil && nextJSCmd.Process != nil {
			nextJSCmd.Process.Kill()
		}
		process.RemoveServerPID()
		os.Exit(0)
	}()

	if !daemonMode {
		fmt.Printf("Dashboard running at http://localhost:%s\n", process.ServerPort)
		fmt.Println("Press Ctrl+C to stop")
	}

	if err := http.ListenAndServe(":"+process.ServerPort, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
		process.RemoveServerPID()
		os.Exit(1)
	}
}

func ensureNextJSReady(nodePath, webPath string) error {
	npmPath, err := runtime.NpmPath()
	if err != nil {
		return err
	}

	nodeModules := filepath.Join(webPath, "node_modules")
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		if !daemonMode {
			fmt.Println("Installing dependencies...")
		}
		cmd := exec.Command(npmPath, "install")
		cmd.Dir = webPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), "PATH="+filepath.Dir(nodePath)+":"+os.Getenv("PATH"))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npm install failed: %w", err)
		}
	}

	buildPath := filepath.Join(webPath, ".next")
	buildInfo, err := os.Stat(buildPath)
	needsBuild := os.IsNotExist(err) || time.Since(buildInfo.ModTime()) > 24*time.Hour

	if needsBuild {
		if !daemonMode {
			fmt.Println("Building Next.js app...")
		}
		cmd := exec.Command(npmPath, "run", "build")
		cmd.Dir = webPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), "PATH="+filepath.Dir(nodePath)+":"+os.Getenv("PATH"))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npm build failed: %w", err)
		}
	}

	return nil
}

func startNextJS(ctx context.Context, nodePath, webPath string) *exec.Cmd {
	npmPath, _ := runtime.NpmPath()

	cmd := exec.CommandContext(ctx, npmPath, "run", "start")
	cmd.Dir = webPath
	cmd.Env = append(os.Environ(), "PORT="+process.NextJSPort, "PATH="+filepath.Dir(nodePath)+":"+os.Getenv("PATH"))

	if daemonMode {
		cmd.Stdout = nil
		cmd.Stderr = nil
	}

	if err := cmd.Start(); err != nil {
		return nil
	}

	return cmd
}

func waitForNextJS(port string, timeout time.Duration) bool {
	checkURL := fmt.Sprintf("http://localhost:%s", port)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := http.Get(checkURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return true
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return false
}

func createUnifiedServer(database *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	nextURL, _ := url.Parse(fmt.Sprintf("http://localhost:%s", process.NextJSPort))
	proxy := httputil.NewSingleHostReverseProxy(nextURL)

	server.RegisterRoutes(mux, database)

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	return mux
}
