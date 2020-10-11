package main

import (
	"fmt"
	"net/url"
	"odbc/db"
	"odbc/rest"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// our clean up procedure and exiting the program.
func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ciao ci vediamo...")
		os.Exit(0)
	}()
}

func main() {

	config, restConfig := rest.LoadConfig()

	rest.UpdateDSN(restConfig.UUID, config.DB, config.Driver)

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
			fmt.Println("panic occurred:", err)
		}
	}()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {

		case <-ticker.C:
			connected, inProgress := rest.ConnectApiStatus()
			if !connected {
				fmt.Println("Not connected")
			}

			if !connected && !inProgress {
				rest.ConnectApi(baseUrl, restConfig)
			}

		case <-interrupt:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				fmt.Println("write close:", err)
			}
			select {

			case <-time.After(time.Second * 5):
			}
			return
		}
	}
}
