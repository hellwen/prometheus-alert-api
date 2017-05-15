package main

import (
	"os"
	"log"
	"strings"
	"net/http"
	"io/ioutil"
	"time"
	"fmt"
	"regexp"
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

func localTime(sourceTime string) (string) {
	loc, _:= time.LoadLocation("Asia/Shanghai")
	f1 := "2006-01-02 15:04:05 Mon"
	t, _ := time.Parse(time.RFC3339, sourceTime)
	return t.In(loc).Format(f1)
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
        // msg_url = "https://oapi.dingtalk.com/robot/send?access_token=f3e4ba2d28e282518b44eafb45761e1f237f70f2af4cc77db4cfde838101f5a5"
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

	// log.Printf("body: %s", jsbody)

	msg := ""

	status := js.Get("status").MustString()
	// receiver := js.Get("receiver").MustString()
	groupLabels, _ := js.Get("groupLabels").Map()
	// commonLabels, _ := js.Get("commonLabels").Map()
	commonAnnotations, _ := js.Get("commonAnnotations").Map()

	eurl_re, _ := regexp.Compile("http://alertmanager.*9093")

	externalURL := js.Get("externalURL").MustString()
	externalURL = eurl_re.ReplaceAllString(externalURL, "http://k8s.gz.1253104200.clb.myqcloud.com:32013")

	log.Printf("request data:")
	log.Printf("groupLabels: %v", groupLabels)

	commonAnnotation := ""
	for k, v := range commonAnnotations {
		if commonAnnotation != "" {
       			commonAnnotation = fmt.Sprintf("%v\n%v: %v", commonAnnotation, k, v)
		} else {
        		commonAnnotation = fmt.Sprintf("%v: %v", k, v)
		}
	}

	/*
	commonLabel := ""
	for k, v := range commonLabels {
		if commonLabel != "" {
       			commonLabel = fmt.Sprintf("%v\n %v: [%v]", commonLabel, k, v)
		} else {
        		commonLabel = fmt.Sprintf(" %v: [%v]", k, v)
		}
	}

	msg = fmt.Sprintf("[%v]\n%v\nLabels:\n%v", status, commonAnnotation, commonLabel)
	*/

	msg = fmt.Sprintf("# [%v] [view](%v)\n%v\n", status, externalURL, commonAnnotation)

	gurl_re, _ := regexp.Compile("http://prometheus.*9090")

	msg = fmt.Sprintf("%v\n# Detail", msg)
	alerts, _ := js.Get("alerts").Array()
	
	for i, a := range alerts {
		na, _ := a.(map[string]interface{})  

		/*
		label := ""
		labels, _ := na["labels"].(map[string]interface{})
		for k, v := range labels {
			if label != "" {
       				label = fmt.Sprintf("%v\n %v: [%v]", label, k, v)
			} else {
       				label = fmt.Sprintf(" %v: [%v]", k, v)
			}
		}
		log.Printf("label: %v", label)
		*/

		annotations := na["annotations"].(map[string]interface{})
		annotation := fmt.Sprintf("%v", annotations["description"])
		// log.Printf("annotation: %v", annotation)

		generatorURL := fmt.Sprintf("%v", na["generatorURL"])
		generatorURL = gurl_re.ReplaceAllString(generatorURL, "http://k8s.gz.1253104200.clb.myqcloud.com:32012")

		startsAt := fmt.Sprintf("%v", na["startsAt"])
		startsAt_local := localTime(startsAt)

		if len(alerts) > 1 {
			msg = fmt.Sprintf("%v\n > %v) \n>%v\n>startsAt: %v\n>[view](%v)", msg, i, annotation, startsAt_local, generatorURL)
		} else {
			msg = fmt.Sprintf("%v\nstartsAt: %v\n[view](%v)", msg, annotation, startsAt_local, generatorURL)
		}
	}

	title := "kubernetes alert"
	msg = fmt.Sprintf("{\"msgtype\":\"markdown\",\"markdown\":{\"title\": \"%v\",\"text\":\"%v\"}}", title, msg)
	return msg
}

func httpPost(tos string, url string, content string) ([]byte, error) {
	// resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader("tos=" + tos + "&content=" + content))
	resp, err := http.Post(url, "application/json", strings.NewReader(content))

	if err != nil {
		return nil, err
	}
 
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func main() {
        m.RunOnAddr(":80")
}
