package main

import (
	"fmt"
	"log"
	"net/url"
	"odbc/db"
	"odbc/rest"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// var watcher *fsnotify.Watcher

// watchDir gets run as a walk func, searching for directories to add watchers to
/* func watchDir(path string, fi os.FileInfo, err error) error {

	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	log.Println(path)
	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}

	return nil
} */

// our clean up procedure and exiting the program.
func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("\r- Ciao ci vediamo...")
		os.Exit(0)
	}()
}

func main() {
	// creates a new file watcher
	//watcher, _ = fsnotify.NewWatcher()
	//defer watcher.Close()

	f, err := os.OpenFile("odbc-welcome.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()
	log.SetOutput(f)

	config, restConfig := rest.LoadConfig()

	rest.UpdateDSN(restConfig.UUID, config.DB, config.Driver)

	log.Println("DATABASE ", config.DB)

	/*
		if err := filepath.Walk(config.DB, watchDir); err != nil {
			log.Println("ERROR", err)
		}

		 	go func() {
			for {
				select {
				// watch for events
				case event := <-watcher.Events:
					log.Printf("Watcher: %s\n", event.Name)
					rest.Watcher(event.Name, restConfig)
					// watch for errors
				case err := <-watcher.Errors:
					log.Println("WATCHER ERROR", err)
				}
			}
		}() */

	SetupCloseHandler()

	if restConfig.WssUrl == "" {
		fmt.Println("No api server defined")
		return
	}

	baseUrl, err := url.Parse(restConfig.WssUrl)
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		return
	}
	if restConfig.UUID == "" {
		fmt.Println("No agent defined")
		return
	}

	fmt.Printf("Agent: %s\nAPI Server: %s\n", restConfig.UUID, baseUrl)

	db.OpenDB(restConfig.UUID)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c := rest.ConnectApi(baseUrl, restConfig)
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {

		case <-ticker.C:
			connected, inProgress := rest.ConnectApiStatus()
			if !connected {
				log.Println("Not connected")
			}

			if !connected && !inProgress {
				rest.ConnectApi(baseUrl, restConfig)
			}

		case <-interrupt:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
			}
			select {

			case <-time.After(time.Second * 5):
			}

			return
		}

	}

}
