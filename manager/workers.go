// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package manager

import (
	"bytes"
	"errors"
	"net"
	"net/http"

	"github.com/gorilla/rpc/v2/json2"
	"github.com/zhanglongx/Aqua/comm"
	"github.com/zhanglongx/Aqua/driver"
)

// Workers store all workers registered by cards
type Workers []driver.Worker

type regInfo struct {
	slot int

	name string

	ip net.IP
}

var (
	errNoCardFound = errors.New("no cards found")
)

// register accept sub-card's register
func (ws *Workers) register(need []string) error {

	var cards []regInfo
	var err error
	if cards, err = onlineCards(); err != nil {
		return err
	}

	cards = append(cards, regInfo{32, "local_encoder", net.IPv4(192, 165, 53, 35)})
	cards = append(cards, regInfo{33, "local_decoder", net.IPv4(192, 165, 53, 35)})

	// FIXME: should be shared between path
	alloced := make(map[int]bool)

	for _, found := range cards {
		inNeed := false
		for _, n := range need {
			if n == found.name {
				inNeed = true
				break
			}
		}

		if inNeed == false {
			continue
		}

		var card driver.Card
		switch found.name {
		case driver.LocalEncoderName:
			card = &driver.LocalE{}
		case driver.LocalDecoderName:
			card = &driver.LocalD{}
		default:
			comm.Error.Printf("Unknown card type %s", found.name)
			continue
		}

		comm.Info.Printf("Registering card %s in slot %d %v",
			found.name, found.slot, found.ip)

		if alloced[found.slot] == true {
			comm.Error.Printf("Slot %d already registered", found.slot)
			continue
		}

		if w, err := card.Open(found.slot, found.ip); err == nil {
			*ws = append(*ws, w...)
			alloced[found.slot] = true
		} else {
			comm.Error.Printf("Open card %s failed", found.name)
		}
	}

	if len(alloced) == 0 {
		return errNoCardFound
	}

	return nil
}

// findWorker finds a worker by worker's name
func (ws *Workers) findWorker(name string) driver.Worker {

	for _, w := range *ws {
		if name == driver.GetWorkerName(w) {
			return w
		}
	}

	return nil
}

func onlineCards() ([]regInfo, error) {

	args := make(map[string]interface{})
	args["cards"] = [0]int{}

	var message []byte
	var err error
	if message, err = json2.EncodeClientRequest("register_server.query", args); err != nil {
		comm.Error.Panicf("%v", err)
	}

	var resp *http.Response
	if resp, err = http.Post(driver.TransURL, "application/json", bytes.NewReader(message)); err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	reply := make(map[string]interface{})
	err = json2.DecodeClientResponse(resp.Body, &reply)
	if err != nil {
		return nil, err
	}

	var result []regInfo

	var v interface{}
	for _, v = range reply["cards"].([]interface{}) {
		name := v.(map[string]interface{})["name"].(string)
		slot := int(v.(map[string]interface{})["slot"].(float64))

		cpus := v.(map[string]interface{})["cpus"].([]interface{})[0]
		ip := net.ParseIP(cpus.(map[string]interface{})["ip"].(string))

		result = append(result, regInfo{slot: slot, name: name, ip: ip})
	}

	return result, nil
}
