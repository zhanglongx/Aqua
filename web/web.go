// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package web

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/rpc/v2/json2"
	"github.com/zhanglongx/Aqua/comm"
	"github.com/zhanglongx/Aqua/manager"
)

// M is shortcut for map
type M map[string]interface{}

// pointers to Path
var (
	ep = &manager.Path{}
	dp = &manager.Path{}
)

// StartAPP launch Web App
func StartAPP() {

	if err := ep.Create(comm.AppCfg.EPDir, comm.AppCfg.EPFile,
		comm.AppCfg.EPNeed); err != nil {
		comm.Error.Panicf("Create EncodePath failed: %v", err)
	}

	if err := dp.Create(comm.AppCfg.DPDir, comm.AppCfg.DPFile,
		comm.AppCfg.DPNeed); err != nil {
		comm.Error.Panicf("Create DecodePath failed: %v", err)
	}
	http.HandleFunc("/encode", encodeIdx)
	http.HandleFunc("/decode", decodeIdx)
	http.HandleFunc("/goform/form_data", jsonrpcIdx)

	if comm.AppCfg.IsHTTPPipeOn {
		http.HandleFunc("/Pipe", pipeIdx)
	}

	stop := make(chan struct{}) // close(stop) will inform most goroutines to stop
	fin := make(chan struct{})  // finish signal of `monitorStatus`

	go monitorStatus(stop, fin, ep.GetWorkers())

	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))

	close(stop) // inform goroutines to stop
	<-fin       // wait for `monitorStatus` to stop
}

func monitorStatus(stop <-chan struct{}, fin chan<- struct{}, workers []string) {
	defer func() { fin <- struct{}{} }()
	var wg sync.WaitGroup

	isStoped := func() bool {
		select {
		case <-stop: // when `stop` is closed, trigger this
			return true
		default:
			return false
		}
	}

	// get worker status every 2 seconds
	queryStatus := func(workerName string, stop <-chan struct{}) {
		defer wg.Done()
		tick := time.NewTicker(2 * time.Second)
		for {
			if isStoped() {
				break
			}
			select {
			case <-tick.C:
				ep.UpdateWorkerStatus(workerName)
			}
		}
	}

	// start `queryStatus` for each worker
	for _, w := range workers {
		wg.Add(1)
		go queryStatus(w, stop)
	}

	// block until all queryStatus finish
	wg.Wait()
}

// TODO: to make a unified idx func
func encodeIdx(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	IDStr := r.Form.Get("ID")
	set := r.Form.Get("set")

	var allErr []error
	if set == "设置参数" {
		if err := setEP(r.Form); err != nil {
			allErr = append(allErr, err)
		}
	}

	data := make(map[interface{}]interface{})

	var content M
	var err error
	if content, err = getEP(IDStr); err != nil {
		allErr = append(allErr, err)
	}

	if len(allErr) > 0 {
		content["Error"] = allErr
	}

	data["Content"] = content

	execTpl(w, data, epTpl)
}

func decodeIdx(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	IDStr := r.Form.Get("ID")
	set := r.Form.Get("set")

	var allErr []error
	if set == "设置参数" {
		if err := setDP(r.Form); err != nil {
			allErr = append(allErr, err)
		}
	}

	data := make(map[interface{}]interface{})

	var content M
	var err error
	if content, err = getDP(IDStr); err != nil {
		allErr = append(allErr, err)
	}

	if len(allErr) > 0 {
		content["Error"] = allErr
	}

	data["Content"] = content

	execTpl(w, data, dpTpl)
}

func pipeIdx(w http.ResponseWriter, r *http.Request) {
	manager.GetPipeInfo(w)

	fmt.Fprintf(w, "%v", ep.GetWorkerStatus())
}

// send jsonrpc to corresponding card and return response
// send jsonrpc to http://[url]/goform/form_data?target=xx_x_x
func jsonrpcIdx(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	codec := json2.NewCodec()
	req := codec.NewRequest(r)

	// get `target` worker from Get form
	r.ParseForm()
	target := r.Form.Get("target")

	if target == "" {
		//TODO? jsonrpc to this software
		req.WriteError(w, 32000, fmt.Errorf("TODO? no target"))
		return
	}
	ip := ep.GetWorkerIP(target)
	if ip == nil {
		req.WriteError(w, 32000, fmt.Errorf("no such worker"))
		return
	}

	// send request body to card
	url := fmt.Sprintf("http://%s/goform/form_data", ip)
	resp, err := http.Post(url, "application/json", r.Body)
	if err != nil {
		req.WriteError(w, 32000, err)
	}
	defer resp.Body.Close()

	// copy response to client
	io.Copy(w, resp.Body)
}

func setEP(val url.Values) error {

	IDStr := val.Get("ID")

	if IDStr == "" {
		return nil
	}

	id, _ := strconv.Atoi(IDStr)

	// FIXME: more checks?
	params := make(manager.Params)
	params["PathName"] = val.Get("PathName")
	params["WorkerName"] = val.Get("WorkerName")

	if val.Get("IsRunning") == "1" {
		params["IsRunning"] = true
	} else {
		params["IsRunning"] = false
	}

	card := make(map[string]interface{})
	card["rtsp_url"] = val.Get("rtsp_url")
	card["BitRate"], _ = strconv.Atoi(val.Get("BitRate"))

	params["Card"] = card

	if err := ep.Set(id, params); err != nil {
		comm.Error.Printf("Set path %d failed", id)
		return err
	}

	return nil
}

func getEP(IDStr string) (M, error) {

	content := make(M)

	// default
	content["ID"] = []int{1, 2, 3, 4}
	content["PathName"] = ""
	content["WorkerName"] = ep.GetWorkers()
	content["IsRunning"] = false

	card := make(M)
	card["rtsp_url"] = ""
	card["BitRate"] = 0

	content["Card"] = card

	if IDStr == "" {
		return content, nil
	}

	id, _ := strconv.Atoi(IDStr)

	var params manager.Params
	var err error
	if params, err = ep.Get(id); err != nil {
		comm.Error.Printf("Get path %d failed", id)
		return content, err
	}

	content["ID"] = selectInt(content["ID"].([]int), id)
	content["PathName"] = params["PathName"]
	content["WorkerName"] = selectStr(content["WorkerName"].([]string),
		params["WorkerName"].(string))
	content["IsRunning"] = params["IsRunning"]

	content["Card"] = params["Card"]

	return content, nil
}

func setDP(val url.Values) error {

	IDStr := val.Get("ID")

	if IDStr == "" {
		return nil
	}

	id, _ := strconv.Atoi(IDStr)

	// FIXME: more checks?
	params := make(manager.Params)
	params["WorkerName"] = val.Get("WorkerName")
	if val.Get("IsRunning") == "1" {
		params["IsRunning"] = true
	} else {
		params["IsRunning"] = false
	}

	if err := dp.Set(id, params); err != nil {
		comm.Error.Printf("Set path %d failed: %s", id, err)
		return err
	}

	return nil
}

func getDP(IDStr string) (M, error) {

	content := make(M)

	// default
	content["ID"] = []int{1, 2, 3, 4}
	content["WorkerName"] = dp.GetWorkers()
	content["IsRunning"] = false

	if IDStr == "" {
		return content, nil
	}

	id, _ := strconv.Atoi(IDStr)

	var params manager.Params
	var err error
	if params, err = dp.Get(id); err != nil {
		comm.Error.Printf("Get path %d failed", id)
		return content, err
	}

	content["ID"] = selectInt(content["ID"].([]int), id)
	content["WorkerName"] = selectStr(content["WorkerName"].([]string),
		params["WorkerName"].(string))
	content["IsRunning"] = params["IsRunning"]

	return content, nil
}

// beego: https://github.com/astaxie/beego
func execTpl(rw http.ResponseWriter, data map[interface{}]interface{}, tpls ...string) {
	tmpl := template.Must(template.New("main").Parse(mainTpl))
	for _, tpl := range tpls {
		tmpl = template.Must(tmpl.Parse(tpl))
	}
	tmpl.Execute(rw, data)
}

func selectStr(list []string, s string) []string {
	var k int
	var l string
	for k, l = range list {
		if l == s {
			ret := []string{l}
			ret = append(ret, list[0:k]...)
			ret = append(ret, list[k+1:]...)

			return ret
		}
	}

	return list
}

func selectInt(list []int, s int) []int {
	var k int
	var l int
	for k, l = range list {
		if l == s {
			ret := []int{l}
			ret = append(ret, list[0:k]...)
			ret = append(ret, list[k+1:]...)

			return ret
		}
	}

	return list
}
