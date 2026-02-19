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
	"github.com/ndzuma/probeTool/internal/server"
	"github.com/spf13/cobra"
)

const (
	ServerPort = "37330"
	NextJSPort = "37331"
	WebDir     = "web"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the probe dashboard server",
	Long:  `Starts both the API server and Next.js frontend on a single port.`,
	Run: func(cmd *cobra.Command, args []string) {
		runServe()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe() {
	fmt.Println("🚀 Starting Probe Dashboard...")

	database, err := db.InitDB(db.DBPath())
	if err != nil {
		fmt.Printf("❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	webPath := filepath.Join(".", WebDir)
	if err := ensureNextJSReady(webPath); err != nil {
		fmt.Printf("❌ Error setting up Next.js: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nextJSCmd := startNextJS(ctx, webPath)
	if nextJSCmd == nil {
		fmt.Println("❌ Failed to start Next.js server")
		os.Exit(1)
	}
	defer func() {
		if nextJSCmd.Process != nil {
			nextJSCmd.Process.Kill()
		}
	}()

	if !waitForNextJS(NextJSPort, 30*time.Second) {
		fmt.Println("❌ Next.js server failed to start")
		os.Exit(1)
	}

	mux := createUnifiedServer(database)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down gracefully...")
		cancel()
		if nextJSCmd != nil && nextJSCmd.Process != nil {
			nextJSCmd.Process.Kill()
		}
		os.Exit(0)
	}()

	addr := ":" + ServerPort
	fmt.Printf("\n✅ Dashboard running at http://localhost:%s\n", ServerPort)
	fmt.Println("   Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
		os.Exit(1)
	}
}

func ensureNextJSReady(webPath string) error {
	buildPath := filepath.Join(webPath, ".next")

	nodeModules := filepath.Join(webPath, "node_modules")
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		fmt.Println("📦 Installing dependencies...")
		cmd := exec.Command("npm", "install")
		cmd.Dir = webPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npm install failed: %w", err)
		}
	}

	buildInfo, err := os.Stat(buildPath)
	needsBuild := os.IsNotExist(err) || time.Since(buildInfo.ModTime()) > 24*time.Hour

	if needsBuild {
		fmt.Println("🔨 Building Next.js app...")
		cmd := exec.Command("npm", "run", "build")
		cmd.Dir = webPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npm build failed: %w", err)
		}
	}

	return nil
}

func startNextJS(ctx context.Context, webPath string) *exec.Cmd {
	fmt.Println("🌐 Starting Next.js server...")

	cmd := exec.CommandContext(ctx, "npm", "run", "start")
	cmd.Dir = webPath
	cmd.Env = append(os.Environ(), "PORT="+NextJSPort)

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start Next.js: %v\n", err)
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

	nextURL, _ := url.Parse(fmt.Sprintf("http://localhost:%s", NextJSPort))
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
