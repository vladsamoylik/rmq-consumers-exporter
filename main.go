package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// how to use flags: https://www.callicoder.com/go-command-line-flags-options-exmple/
var version = "1.0"
var listen_port = flag.String("listen-port", ":9201", "specify listen port for exporter")
var url = flag.String("url", "http://127.0.0.1:15672", "specify your rmq api url")
var username = flag.String("username", "user", "specify user for rmq authentication")
var password = flag.String("password", "passwd", "specify password for rmq authentication")
var consumer = flag.String("consumer", "ecallmgr_fs_conferences_shared", "specify consumer name in rmq")

// go struct from json: https://mholt.github.io/json-to-go/
type Response struct {
	Items []struct {
		Consumers int    `json:"consumers"`
		Name      string `json:"name"`
		Node      string `json:"node"`
	}
}

func makeUri() string {
	flag.Parse()
	api_uri_merge := []string{*url, "/api/queues?page=1", "&name=", *consumer}
	api_uri := strings.Join(api_uri_merge, "")
	return api_uri
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to RabbitMQ consumers exporter", "version: ", version)
	fmt.Fprintln(w, "The metrics are available here: /metrics ")
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	programName := filepath.Base(os.Args[0])
	sysLog, err := syslog.New(syslog.LOG_INFO|syslog.LOG_LOCAL7, programName)
	if err != nil {
		log.Println(err)
		return
	} else {
		log.SetOutput(sysLog)
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", makeUri(), nil)
	if err != nil {
		log.Println(err)
		return
	} else {
		log.SetOutput(sysLog)
	}

	flag.Parse()
	req.SetBasicAuth(*username, *password)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	} else {
		log.SetOutput(sysLog)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	} else {
		log.SetOutput(sysLog)
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}

	for _, rec := range result.Items {
		consumerMetric := fmt.Sprintf("ecallmgr_fs_conferences_shared{conference_name=\"%s\", node_name=\"%s\"} %d", rec.Name, rec.Node, rec.Consumers)
		fmt.Fprintf(w, "%s\n", consumerMetric)
	}

}

func main() {
	flag.Parse()
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/metrics", metricsHandler)
	http.ListenAndServe(*listen_port, nil)
	err := http.ListenAndServe(*listen_port, nil)
	if err != nil {
		log.Println(err)
		return
	}
}
