package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "os"
    "strings"
    "sync/atomic"
)

// Global counter for the total number of requests.
var requestCount int64

func main() {
    // Configure log file for output
    logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("error opening log file: %v", err)
    }
    defer logFile.Close()

    // MultiWriter to direct logs both to standard output and the log file.
    mw := io.MultiWriter(os.Stdout, logFile)
    log.SetOutput(mw)

    // HTTP Handlers
    // Handler for favicon.ico requests
    http.HandleFunc("/favicon.ico", handleFavicon)

    // Main handler for root path
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Increment counter only for GET requests to the root path.
        if r.Method == http.MethodGet && r.URL.Path == "/" {
            atomic.AddInt64(&requestCount, 1)
        }

        deploymentColor := getDefaultEnv("DEPLOYMENT_COLOR", "black")
        deploymentBackground := getDefaultEnv("DEPLOYMENT_BACKGROUND", "white")
        hostname, err := os.Hostname()
        if err != nil {
            log.Printf("Failed to get hostname: %v", err)
            fmt.Fprintf(w, "Failed to get hostname: %v", err)
            return
        }

        clientIP := getClientIP(r)
        deploymentVersion := fmt.Sprintf("This is the %s deployment running on %s. Total page requests: %d", deploymentColor, hostname, requestCount)
        clientInfo := fmt.Sprintf("Visitor IP: %s", clientIP)

        log.Printf("Page accessed by %s", clientIP)

        // Generate and send the HTML response
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprintf(w, generateHTML(deploymentColor, deploymentBackground, deploymentVersion, clientInfo))
    })

    // Start the HTTP server
    log.Println("Starting server...")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

// handleFavicon is a dedicated handler for favicon.ico requests
func handleFavicon(w http.ResponseWriter, r *http.Request) {
    // Explicitly ignore or alternatively serve a specific favicon
    http.ServeFile(w, r, "/favicon.ico")
}

// Helper function to safely retrieve environment variables with a fallback
func getDefaultEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}

func getClientIP(r *http.Request) string {
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        log.Printf("Unable to parse remote address '%v': %v", r.RemoteAddr, err)
        return ""
    }

    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        ip = strings.Split(forwarded, ",")[0]
    }
    return ip
}

func generateHTML(deploymentColor, deploymentBackground, deploymentVersion, clientInfo string) string {
    return fmt.Sprintf(`
<html>
<head>
    <title>Deployment Page</title>
    <style>
        body { color: %s; background-color: %s; font-family: 'Arial', sans-serif;
               display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
        .container { text-align: center; border: 1px solid #ddd; box-shadow: 0 4px 6px rgba(0,0,0,0.1); padding: 20px; 
                     border-radius: 8px; max-width: 500px; }
        .footer { margin-top: 20px; font-size: 0.8em; color: #666; }
    </style>
    <script>
        setTimeout(function(){ window.location.reload(1); }, 1000);
    </script>
</head>
<body>
    <div class="container">
        <h1>Deployment Info</h1>
        <p>%s</p>
        <p>%s</p>
        <div class="footer">Powered by Go</div>
    </div>
</body>
</html>`, deploymentColor, deploymentBackground, deploymentVersion, clientInfo)
}