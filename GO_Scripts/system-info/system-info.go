package main

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"time"
)

type SystemInfo struct {
	OS           string    `json:"os"`
	Architecture string    `json:"architecture"`
	CPUs         int       `json:"cpus"`
	GoVersion    string    `json:"go_version"`
	Hostname     string    `json:"hostname"`
	Username     string    `json:"username"`
	HomeDir      string    `json:"home_dir"`
	WorkingDir   string    `json:"working_dir"`
	Timestamp    time.Time `json:"timestamp"`
}

func GetSystemInfo() (*SystemInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "unknown"
	}

	return &SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		CPUs:         runtime.NumCPU(),
		GoVersion:    runtime.Version(),
		Hostname:     hostname,
		Username:     currentUser.Username,
		HomeDir:      currentUser.HomeDir,
		WorkingDir:   workingDir,
		Timestamp:    time.Now(),
	}, nil
}

func GetEnvironmentVariables() map[string]string {
	envVars := make(map[string]string)
	
	for _, env := range os.Environ() {
		// Split on first '=' only
		for i, char := range env {
			if char == '=' {
				key := env[:i]
				value := env[i+1:]
				envVars[key] = value
				break
			}
		}
	}
	
	return envVars
}

func GetMemoryStats() runtime.MemStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats
}

func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	if bytes >= GB {
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	} else if bytes >= MB {
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	} else if bytes >= KB {
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	}
	return fmt.Sprintf("%d bytes", bytes)
}

func PrintSystemInfo(info *SystemInfo) {
	fmt.Println("=== System Information ===")
	fmt.Printf("Operating System: %s\n", info.OS)
	fmt.Printf("Architecture: %s\n", info.Architecture)
	fmt.Printf("CPU Cores: %d\n", info.CPUs)
	fmt.Printf("Go Version: %s\n", info.GoVersion)
	fmt.Printf("Hostname: %s\n", info.Hostname)
	fmt.Printf("Username: %s\n", info.Username)
	fmt.Printf("Home Directory: %s\n", info.HomeDir)
	fmt.Printf("Working Directory: %s\n", info.WorkingDir)
	fmt.Printf("Timestamp: %s\n", info.Timestamp.Format("2006-01-02 15:04:05"))
}

func PrintMemoryStats() {
	fmt.Println("\n=== Memory Statistics ===")
	memStats := GetMemoryStats()
	
	fmt.Printf("Allocated Memory: %s\n", FormatBytes(memStats.Alloc))
	fmt.Printf("Total Allocated: %s\n", FormatBytes(memStats.TotalAlloc))
	fmt.Printf("System Memory: %s\n", FormatBytes(memStats.Sys))
	fmt.Printf("Number of GC Cycles: %d\n", memStats.NumGC)
	fmt.Printf("GC CPU Fraction: %.4f\n", memStats.GCCPUFraction)
	fmt.Printf("Heap Objects: %d\n", memStats.HeapObjects)
	fmt.Printf("Heap Size: %s\n", FormatBytes(memStats.HeapSys))
	fmt.Printf("Heap In Use: %s\n", FormatBytes(memStats.HeapInuse))
	fmt.Printf("Heap Released: %s\n", FormatBytes(memStats.HeapReleased))
}

func PrintEnvironmentVariables() {
	fmt.Println("\n=== Environment Variables ===")
	envVars := GetEnvironmentVariables()
	
	// Print some common/important environment variables
	important := []string{
		"PATH", "HOME", "USER", "USERNAME", "USERPROFILE",
		"GOPATH", "GOROOT", "GOPROXY", "GO111MODULE",
		"SHELL", "TERM", "LANG", "TZ",
		"TEMP", "TMP", "TMPDIR",
		"PWD", "OLDPWD",
	}
	
	fmt.Println("Important Environment Variables:")
	for _, key := range important {
		if value, exists := envVars[key]; exists {
			// Truncate very long values
			if len(value) > 100 {
				value = value[:97] + "..."
			}
			fmt.Printf("  %s = %s\n", key, value)
		}
	}
	
	fmt.Printf("\nTotal Environment Variables: %d\n", len(envVars))
}

func PrintRuntimeInfo() {
	fmt.Println("\n=== Go Runtime Information ===")
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("Compiler: %s\n", runtime.Compiler)
	fmt.Printf("GOOS: %s\n", runtime.GOOS)
	fmt.Printf("GOARCH: %s\n", runtime.GOARCH)
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())
	fmt.Printf("NumGoroutine: %d\n", runtime.NumGoroutine())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
}

func main() {
	if len(os.Args) > 1 {
		command := os.Args[1]
		
		switch command {
		case "basic":
			info, err := GetSystemInfo()
			if err != nil {
				fmt.Printf("Error getting system info: %v\n", err)
				os.Exit(1)
			}
			PrintSystemInfo(info)
			
		case "memory":
			PrintMemoryStats()
			
		case "env":
			PrintEnvironmentVariables()
			
		case "runtime":
			PrintRuntimeInfo()
			
		case "all":
			info, err := GetSystemInfo()
			if err != nil {
				fmt.Printf("Error getting system info: %v\n", err)
				os.Exit(1)
			}
			PrintSystemInfo(info)
			PrintMemoryStats()
			PrintEnvironmentVariables()
			PrintRuntimeInfo()
			
		case "help":
			fmt.Println("Usage: go run system-info.go [command]")
			fmt.Println("Commands:")
			fmt.Println("  basic    - Basic system information (default)")
			fmt.Println("  memory   - Memory statistics")
			fmt.Println("  env      - Environment variables")
			fmt.Println("  runtime  - Go runtime information")
			fmt.Println("  all      - All information")
			fmt.Println("  help     - Show this help")
			
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Use 'help' for available commands")
			os.Exit(1)
		}
	} else {
		// Default: show basic system info
		info, err := GetSystemInfo()
		if err != nil {
			fmt.Printf("Error getting system info: %v\n", err)
			os.Exit(1)
		}
		PrintSystemInfo(info)
	}
}
