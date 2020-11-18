package rest

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"odbc/db"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

type agentConfig struct {
	Config string
	DB     string
	Driver string
}

type RestConfig struct {
	UUID       string `json:"uuid"`
	OrionId    string `json:"orionId"`
	OrionSocId string `json:"orionSocId"`
	Email      string `json:"email"`
	Societa    string `json:"societa"`
	WssUrl     string `json:"wssUrl"`
}

type command struct {
	Cmd   string
	Data  string
	Agent string
	Id    int
}

type response struct {
	Cmd   string
	Data  map[int]map[string]string
	Agent string
	Id    int
	Error string
}

var (
	apiConnection          *websocket.Conn
	apiConnected           = false
	apiConnectedInProgress = false
	Agent                  = "UnkonwnAgent"
)

func LoadConfig() (agentConfig, RestConfig) {
	config, err := ioutil.ReadFile("./agent.json")
	if err != nil {
		log.Fatal("No agent.json config file")
	}

	var a = agentConfig{}
	if err := json.Unmarshal(config, &a); err != nil {
		panic(err)
	}

	log.Println("Configuration File: ", a.Config)

	abs := filepath.IsAbs(a.DB)
	if abs {
		_, err := os.Stat(a.DB)
		if os.IsNotExist(err) {
			log.Fatal("File " + a.DB + ": Impossibile trovare il file specificato.")
		}
		log.Println("DataBase Dir is ABSOLUTE: ", a.DB)
	} else {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		_, err = os.Stat(dir + `\` + a.DB)
		if os.IsNotExist(err) {
			log.Fatal("File " + dir + `\` + a.DB + ": Impossibile trovare il file specificato.")
		}
		a.DB = dir + `\` + a.DB
		log.Println("DataBase Dir is RELATIVE: ", a.DB)
	}

	config, err = ioutil.ReadFile("./" + a.Config)
	if err != nil {
		log.Fatal("No " + a.Config + "config file")
	}
	var restConfig = RestConfig{}
	if err := json.Unmarshal(config, &restConfig); err != nil {
		panic(err)
	}
	return a, restConfig
}

func ConnectApiStatus() (bool, bool) {
	return apiConnected, apiConnectedInProgress
}

func ConnectApi(api *url.URL, agent RestConfig) *websocket.Conn {
	//*addr
	apiConnectedInProgress = true
	u := url.URL{Scheme: api.Scheme, Host: api.Host, Path: "/odbc"}
	log.Printf("Connessione a %s\n", u.String())
	firstTentative := true
	for {
		if firstTentative {
			firstTentative = false
			log.Println("Connessione... ", u.String())
		} else {
			log.Println("Riconnessione... ", u.String())
		}

		var err error
		apiConnection, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			apiConnected = false
			log.Println("Impossibile stabilire la connessione. Rifiuto persistente del computer di destinazione.")
			time.Sleep(4 * time.Second)
			continue
		}

		log.Println("API Connesso")
		break
	}
	apiConnected = true
	apiConnectedInProgress = false

	Boot(agent)
	Process()

	return apiConnection
}

func Boot(agent RestConfig) {
	b, err := json.Marshal(agent)
	if err != nil {
		log.Println("error:", err)
	}
	m := command{
		Cmd:   "boot",
		Data:  string(b),
		Id:    0,
		Agent: agent.UUID,
	}
	j, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
	}

	err = apiConnection.WriteMessage(websocket.TextMessage, j)
	if err != nil {
		log.Println("Errore scrittura: ", err)
	}
}

func Watcher(file string, agent RestConfig) {
	m := command{
		Cmd:   "watcher",
		Data:  file,
		Id:    0,
		Agent: agent.UUID,
	}
	j, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
	}

	err = apiConnection.WriteMessage(websocket.TextMessage, j)
	if err != nil {
		log.Println("Errore scrittura: ", err)
	}
}

func Process() {
	done := make(chan struct{})
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Errore di processo:", err)
				apiConnected = false
				apiConnection.Close()
			}
		}()
		defer close(done)
		for {
			if apiConnected {
				_, message, err := apiConnection.ReadMessage()
				if err != nil {
					log.Println("Lettura WS:", err)
				}
				m := command{}
				if err := json.Unmarshal(message, &m); err != nil {
					log.Println(err)
				}
				if m.Cmd == "query" {
					log.Printf("%s [%d]\nSQL %s\n", m.Agent, m.Id, m.Data)
					response := response{
						Cmd:   m.Cmd,
						Id:    m.Id,
						Agent: m.Agent,
					}
					data, err := db.SqlQuery(m.Data)
					if err != nil {
						response.Error = err.Error()
					}

					response.Data = data

					j, err := json.Marshal(response)
					if err != nil {
						log.Println("Errore assemblaggio JSON:", err)
					}
					log.Printf("Lunghezza dati: %d\n------\n", len(data))
					err = apiConnection.WriteMessage(websocket.TextMessage, j)
					if err != nil {
						log.Println("Scrittura WS:", err)
					}
				}
			}
		}
	}()
}
