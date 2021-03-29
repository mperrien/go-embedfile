package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	
	"github.com/pkg/browser"
)


func main() {
	httpServerWaitGroup := &sync.WaitGroup{}

	// This will allow to gracefully shutdown server
	// See https://medium.com/honestbee-tw-engineer/gracefully-shutdown-in-go-http-server-5f5e6b83da5a
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	httpServerWaitGroup.Add(2) // Adding 2 go routines
	srv := startServer(httpServerWaitGroup)
	// log.Print("Server Started")
	go openBrowser(httpServerWaitGroup)
	
	fmt.Println("Press Enter to close server when you're done!")
	fmt.Println("(It won't work, you need to CTRL+C to exit.)")
  // fmt.Scanln()

	<-done
	// log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// Close database, redis, truncate message queues, etc.
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	fmt.Printf("\nServer Exited Properly\n")
	
	httpServerWaitGroup.Wait()
	
}

//go:embed reportfiles
var reportFiles embed.FS

func getReportFiles() http.FileSystem {
	rfiles, err := fs.Sub(reportFiles, "reportfiles")
	if err != nil {
			panic(err)
	}

	return http.FS(rfiles)
}

func startServer(wg *sync.WaitGroup) *http.Server {
	srv := &http.Server{Addr: ":8888"}
	http.Handle("/", http.FileServer(getReportFiles()))

	go func(){
		defer wg.Done()

		// always returns error. ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// unexpected error. port in use?
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()
	
	// returning reference so caller can call Shutdown()
	return srv
	
}

func openBrowser(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Your report is available at http://localhost:8888.")
	fmt.Println("Opening browser...")
	const reportUrl = "http://localhost:8888/"
	browser.OpenURL(reportUrl)
}