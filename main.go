package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "os"
    "strings"
)

func main() {
    // Open or create the log file
    logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("error opening log file: %v", err)
    }
    defer logFile.Close()

    // Set log output to both console and file
    mw := io.MultiWriter(os.Stdout, logFile)
    log.SetOutput(mw)

    // Setup the HTTP server
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        deploymentColor := os.Getenv("DEPLOYMENT_COLOR")
        deploymentBackground := os.Getenv("DEPLOYMENT_BACKGROUND")

        if deploymentColor == "" {
            deploymentColor = "black"
        }
        if deploymentBackground == "" {
            deploymentBackground = "white"
        }

        hostname, err := os.Hostname()
        if err != nil {
            log.Printf("Failed to get hostname: %v", err)
            fmt.Fprintf(w, "Failed to get hostname: %v", err)
            return
        }

        clientIP := getClientIP(r)
        deploymentVersion := fmt.Sprintf("This is the %s deployment running on %s", deploymentColor, hostname)
        clientInfo := fmt.Sprintf("Visitor IP: %s", clientIP)

        log.Printf("Page accessed by %s", clientIP)

        // Generate and send the HTML response
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprintf(w, generateHTML(deploymentColor, deploymentBackground, deploymentVersion, clientInfo))
    })

    // Start the HTTP server
    port := "8080"
    log.Printf("Starting server at port %s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func getClientIP(r *http.Request) string {
    // Direct connection, no proxy
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        // Unable to split remote address into IP:Port
        log.Printf("Unable to parse remote address '%v': %v", r.RemoteAddr, err)
        return ""
    }

    // X-Forwarded-For HTTP header decomposition to fetch initial requester's address
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        // Take the first IP in the chain as the original sender
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
            body { color: %s; background-color: %s;
            font-family: 'Arial', sans-serif;
            display: flex; justify-content: center; align-items: center;
            height: 100vh; margin: 0; }
            .container { text-align: center; border: 1px solid #ddd;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1); padding: 20px;
            border-radius: 8px; max-width: 500px; }
            .footer { margin-top: 20px; font-size: 0.8em; color: #666; }
        </style>
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
