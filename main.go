package main

import (
	"os"
	"log"
	"strings"
	"net/http"
	"io/ioutil"
	"time"
	"fmt"
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
		r.Post("/:id", alertID)
	})
	m.Action(r.Handle)
}

func alertDBA(r *http.Request) {
	log.Printf("alert to dba...")
	alert(r, "6")
}

func alertOPS(r *http.Request) {
        log.Printf("alert to ops...")
	alert(r, "5")
}

func alertID(r *http.Request, params martini.Params) {
	alert_id := params["id"]

	log.Printf("alert to %v...", alert_id)
	alert(r, alert_id)
}

func alert(r *http.Request, alert_id string) {
	message := prometheusMessage(r)

        loc, _:= time.LoadLocation("Asia/Shanghai")
        f1 := "2006-01-02 15:04:05 Mon"
        message = fmt.Sprintf("%v at %v", message, time.Now().In(loc).Format(f1))

	log.Printf("message: %s", message)

	msg_url := getMsgUrl()
	resp, err := httpPost(alert_id, msg_url, message)
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

func prometheusMessage(r *http.Request) string {
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

	msg := ""
	// msg = msg + js.Get("externalURL").MustString()

	alerts, _ := js.Get("alerts").Array()
	for _, a := range alerts {
		na, _ := a.(map[string]interface{})  
		labels := na["labels"]
		annotations := na["annotations"]
		startsAt := na["startsAt"]
		endsAt := na["endsAt"]

		msg = msg + "##" + labels + "|" + annotations + "|" + startsAt + "|" + endsAt
	}

	return msg
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
        m.RunOnAddr(":80")
}
