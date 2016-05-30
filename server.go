package main

import (
	"encoding/json"
	"log"
	"net/http"

	"./twilio"
	"./websockets"
	"github.com/gorilla/websocket"
)

var (
	addr      = "localhost:12345"
	bytesChan chan []byte
	hub       websockets.Hub
)

func init() {
	bytesChan = make(chan []byte)
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./client")))
	http.HandleFunc("/sms", handleSMS)
	http.HandleFunc("/ws", handleWS)

	log.Printf("Twilio API running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

/*
	sample parsed form message
	map[From:[+17199635973] FromState:[CO] AccountSid:[ACfed13ffb86b6d0568811a733b98d014b] ApiVersion:[2010-04-01] ToZip:[] AddOns:[{"status":"successful","message":null,"code":null,"results":{"ibm_watson_sentiment":{"request_sid":"XRbd13ad845654ae2c7e4f0b14c7d6a7a3","status":"successful","message":null,"code":null,"result":{"status":"OK","language":"english","docSentiment":{"score":"0.741467","type":"positive"}}}}}] SmsSid:[SMdce1e7018aff73625a0b4671405e6512] ToCountry:[US] ToState:[CO] ToCity:[] SmsMessageSid:[SMdce1e7018aff73625a0b4671405e6512] NumMedia:[0] MessageSid:[SMdce1e7018aff73625a0b4671405e6512] Body:[Really fun to use it for a while and then the other side ] NumSegments:[1] To:[+17192993876] FromZip:[80922] SmsStatus:[received] FromCity:[COLORADO SPRINGS] FromCountry:[US]]
*/

func handleSMS(w http.ResponseWriter, r *http.Request) {
	log.Println("SMS request")

	smsTwiml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
	<Response>
		<Message>
			Thanks for your feedback!
		</Message>
	</Response>`)

	// respond
	_, err := w.Write(smsTwiml)
	if err != nil {
		log.Printf("writer error: %s", err)
	}

	err = r.ParseForm()
	if err != nil {
		log.Printf("parse form error: %s", err)
		return
	}

	ao := twilio.AddOns{}
	b := []byte(r.Form.Get("AddOns"))
	err = json.Unmarshal(b, &ao)
	if err != nil {
		log.Printf("Twilio AddOns parse error: %s", err)
		return
	}

	ds := ao.Results.IBMWatsonSentiment.Result.DocSentiment
	m := twilio.Message{
		Content: r.Form.Get("Body"),
		From:    r.Form.Get("From"),
		Score:   float64(ds.Score),
		Type:    ds.Type,
	}

	b, err = json.Marshal(m)
	if err != nil {
		log.Printf("json marshal error on message: %s", err)
		return
	}

	bytesChan <- b
}

// TODO fix websocket mem leak; like a PingCloser
func handleWS(w http.ResponseWriter, r *http.Request) {
	log.Println("websocket request")

	up := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade websocket error: %s", err)
		return
	}
	defer conn.Close()

	hub.Add(conn)

	for b := range bytesChan {
		log.Printf("sending over websocket: %s", string(b))
		hub.Publish(b)
	}
}
