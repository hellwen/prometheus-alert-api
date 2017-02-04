package main

import (
	"os"
	"log"
	"errors"
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
	r.Post(`/api/alert`, kapacitorAlert)
	// Add the router action
	m.Action(r.Handle)
}

func kapacitorAlert(r *http.Request) {
	log.Printf("alert...")

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

	message := js.Get("message").MustString()
	log.Printf("message: %s", message)

	resp, err := httpPost(message)
	if err != nil {
		log.Println(err)
		log.Println("Message not sended!")
	} else if resp != nil {
		log.Printf("respone: %s", string(resp))
	}
}

func httpPost(content string) ([]byte, error) {
        wechat_url := os.Getenv("WECHAT_URL")

	if wechat_url == "" {
		return nil, errors.New("wechat_url is null! Please set the env var WECHAT_URL")
	}

	resp, err := http.Post(wechat_url, "application/x-www-form-urlencoded", strings.NewReader("tos=2&content=" + content))

	if err != nil {
		return nil, err
	}
 
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func main() {
	m.Run()
}
