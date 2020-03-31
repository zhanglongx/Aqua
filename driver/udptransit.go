// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package driver

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/rpc/v2/json2"
	"github.com/zhanglongx/Aqua/comm"
)

const (
	transURL = "http://localhost/goform/form_data"
)

var (
	errTransitGeneric = errors.New("Transit Generic error")
)

type transit struct {
	lock sync.Mutex

	selfIP net.IP

	seq int
}

func (t *transit) add(srcPort int, dstIP net.IP, dstPort int,
	pair bool) error {

	t.lock.Lock()

	defer t.lock.Unlock()

	num := 1
	if pair == true {
		num = 2
	}

	transponds := make([]map[string]interface{}, num)
	for i := 0; i < num; i++ {

		transponds[0] = make(map[string]interface{})

		transponds[0]["type"] = "udp2udp"
		transponds[0]["recv_ip"] = fmt.Sprintf("%s", t.selfIP)
		transponds[0]["recv_port"] = srcPort + 2*i
		transponds[0]["send_ip"] = fmt.Sprintf("%s", dstIP)
		transponds[0]["send_port"] = dstPort + 2*i

	}

	args := make(map[string]interface{})
	args["transponds"] = transponds

	var message []byte
	var err error
	if message, err = json2.EncodeClientRequest("udp_transpond.add", args); err != nil {
		comm.Error.Panicf("%v", err)
	}

	var resp *http.Response
	if resp, err = http.Post(transURL, "application/json", bytes.NewReader(message)); err != nil {
		return err
	}

	defer resp.Body.Close()

	reply := make(map[string]interface{})
	err = json2.DecodeClientResponse(resp.Body, &reply)
	if err != nil {
		return err
	}

	if reply["status"] != "Established" {
		return errTransitGeneric
	}

	return nil
}

func (t *transit) del(srcPort int, dstIP net.IP, dstPort int,
	pair bool) error {

	t.lock.Lock()

	defer t.lock.Unlock()

	num := 1
	if pair == true {
		num = 2
	}

	transponds := make([]map[string]interface{}, num)
	for i := 0; i < num; i++ {

		transponds[0] = make(map[string]interface{})

		transponds[0]["type"] = "udp2udp"
		transponds[0]["recv_ip"] = fmt.Sprintf("%s", t.selfIP)
		transponds[0]["recv_port"] = srcPort + 2*i
		transponds[0]["send_ip"] = fmt.Sprintf("%s", dstIP)
		transponds[0]["send_port"] = dstPort + 2*i

	}

	args := make(map[string]interface{})
	args["transponds"] = transponds

	var message []byte
	var err error
	if message, err = json2.EncodeClientRequest("udp_transpond.del", args); err != nil {
		comm.Error.Panicf("%v", err)
	}

	var resp *http.Response
	if resp, err = http.Post(transURL, "application/json", bytes.NewReader(message)); err != nil {
		return err
	}

	defer resp.Body.Close()

	reply := make(map[string]interface{})
	err = json2.DecodeClientResponse(resp.Body, &reply)
	if err != nil {
		return err
	}

	if reply["status"] != "Deleted" {
		return errTransitGeneric
	}

	return nil
}
