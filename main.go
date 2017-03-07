package main

import (
	"os"
	"log"
	"strings"
	"net/http"
	"io/ioutil"
	"github.com/go-martini/martini"
	"github.com/bitly/go-simplejson"
)

var m *martini.Martini

func init() {
	m = martini.New()
	// Setup middleware
	m.Use(martini.Recovery())
	m.Use(martini.Logger())
	// Setup routes
	r := martini.NewRouter()
	r.Group("/api/alert", func(r martini.Router) {
		r.Post("/dba", alertDBA)
		r.Post("/ops", alertOPS)
	})
	// r.Post(`/api/alert/ops`, kapacitorAlert)
	// r.Post(`/api/alert/dba`, kapacitorAlert)
	// Add the router action
	m.Action(r.Handle)
}

func alertDBA(r *http.Request) {
	log.Printf("alert to dba...")

        message := kapacitorMessage(r)

	log.Printf("message: %s", message)

	msg_url := getMsgUrl()
	resp, err := httpPost("6", msg_url, message)
	if err != nil {
		log.Println(err)
		log.Println("Message not sended!")
	} else if resp != nil {
		log.Printf("respone: %s", string(resp))
	}
}

func alertOPS(r *http.Request) {
        log.Printf("alert to ops...")

        message := kapacitorMessage(r)

        log.Printf("message: %s", message)

        msg_url := getMsgUrl()
        resp, err := httpPost("5", msg_url, message)
        if err != nil {
                log.Println(err)
                log.Println("Message not sended!")
        } else if resp != nil {
                log.Printf("respone: %s", string(resp))
        }
}

func getMsgUrl() string {
        msg_url := os.Getenv("MSG_URL")
	if msg_url == "" {
		log.Printf("msg_url is null! Please set the env var MSG_URL")
	}

        return msg_url
}

func kapacitorMessage(r *http.Request) string {
	defer r.Body.Close()
	jsbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("body error")
		log.Println(err)
	}

        js, err := simplejson.NewJson(jsbody)
        if err != nil {
		log.Println("json format error")
		log.Println(err)
        }

	// log.Printf("body: %s", body)

	return js.Get("message").MustString()
}

func httpPost(tos string, url string, content string) ([]byte, error) {
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader("tos=" + tos + "&content=" + content))

	if err != nil {
		return nil, err
	}
 
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func main() {
	m.Run()
}
