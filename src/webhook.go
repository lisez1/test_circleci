package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type botClient struct {
	sessionToken string
	kmToken      string
	config
}

type incomingWebhookRequest struct {
	Itemname        string             `json:"item_name"`
	Itemid    string             `json:"item_id"`
	Itemtype     string             `json:"item_type"`
	Itemdes string             `json:"item_description"`
}


func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func (botClient *botClient) postWebhook(w http.ResponseWriter, r *http.Request) {

	clientIP := getIP(r)
	log.Info("webhook got a new request from " + clientIP)

	params := mux.Vars(r)
	var incomingWebhookRequest incomingWebhookRequest

	contentType := r.Header.Get("Content-Type")
	switch strings.Split(contentType, "; ")[0] {
	case "application/x-www-form-urlencoded":
		payload := strings.NewReader(r.FormValue("payload"))
		log.Info("a original message  : " + r.FormValue("payload"))
		err := json.NewDecoder(payload).Decode(&incomingWebhookRequest)
		if err != nil {
			log.Warn(err.Error())
		}
	case "application/json":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Warn(err.Error())
		}
		log.Info("a original message  : " + string(body))
		err = json.Unmarshal(body, &incomingWebhookRequest)
		if err != nil {
			log.Warn(err.Error())
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"an error occured in webhook": "Content-Type should be application/x-www-form-urlencoded or application/json"}`)
	}

	var sendMessageText string

	if incomingWebhookRequest.Itemid != "" {
//		sendMessageText += urlHandler(incomingWebhookRequest.Itemid)
        sendMessageText = "new file uploaded"
	}

	log.Info("the message converted to MLformat : " + sendMessageText)

	responseBody := new(responseBody)
	responseBody, err := sendMessage(botClient, params["streamId"], sendMessageText)
	if err != nil {
		webhookResopse(w, http.StatusBadRequest, err.Error())
	}

	if responseBody.StatusCode != http.StatusOK {
		webhookResopse(w, http.StatusBadRequest, responseBody.Message)
	} else {
		webhookResopse(w, http.StatusOK, responseBody.Message)
	}

}

// getIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func getIP(r *http.Request) string {

	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func webhookResopse(w http.ResponseWriter, respStatus int, respMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respStatus)
	io.WriteString(w, `{"MessageFromSymphony": "`+respMessage+`"}`)
}

func initWebhook(botClient *botClient) {
	router := mux.NewRouter()
	router.HandleFunc("/symphony-hooks/{streamId}", botClient.postWebhook).Methods("POST")
	log.Info("webhook is listening on port 8445.")
	http.ListenAndServe(":8445", router)
}
