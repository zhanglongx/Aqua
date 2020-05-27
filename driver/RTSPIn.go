// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package driver

import (
	"fmt"
	"net"
	"sync"
)

// RTSPInName is the sub-card's name
const RTSPInName string = "RTSPIn"

// RTSPIn is the main struct for sub-card
type RTSPIn struct {
	lock sync.RWMutex

	// Card Slot
	Slot int

	// Card IP
	IP net.IP

	URL string
}

// RTSPInWorker is the main struct for sub-card's
// Worker
type RTSPInWorker struct {
	workerID int

	card *RTSPIn

	rpc map[string]interface{}

	// for status query
	sess Session
}

func newRPC(ip net.IP) map[string]interface{} {

	// XXX: RTSP shared the same IP with udp transit
	return map[string]interface{}{
		"transponds": []interface{}{
			map[string]interface{}{
				"type":     "udp2udp",
				"rtsp_url": "",
				"recv_ip":  ip.String(),
				"send_ip":  ip.String(),
				"send_port": map[string]interface{}{
					"video": 0,
					"audio": 0},
			},
		},
	}
}

// Open method
func (c *RTSPIn) Open() ([]Worker, error) {

	return []Worker{
		&RTSPInWorker{
			workerID: 0,
			card:     c,
			rpc:      newRPC(c.IP),
		},
		&RTSPInWorker{
			workerID: 1,
			card:     c,
			rpc:      newRPC(c.IP),
		},
	}, nil
}

// Close method
func (c *RTSPIn) Close() error {
	return nil
}

// Control method
func (w *RTSPInWorker) Control(c CtlCmd, arg interface{}) interface{} {
	card := w.card

	switch c {
	case CtlCmdStart:
		// leave to CtlCmdSetting

	case CtlCmdStop:
		// leave to CtlCmdSetting

	case CtlCmdName:
		return fmt.Sprintf("%s_%d_%d", C9830TranscoderName,
			card.Slot, w.workerID)

	case CtlCmdIP:
		return card.IP

	case CtlCmdWorkerID:
		return w.workerID

	case CtlCmdSetting:
		if settings, ok := arg.(map[string]interface{}); ok {
			if err := w.set(w.workerID, settings); err != nil {
				return err
			}
		}

	default:
	}
	return nil
}

// Monitor .
func (w *RTSPInWorker) Monitor() (ret bool) {
	// to handle interface conversion error
	defer func() {
		if p := recover(); p != nil {
			// comm.Error.Println(p)
			ret = false
		}
	}()

	ret = true

	params := map[string]interface{}{
		"transponds": []interface{}{
			map[string]interface{}{
				"type":    "udp2udp",
				"recv_ip": "",
				"send_ip": w.sess.IP,
				"send_port": map[string]interface{}{
					"video": w.sess.Ports[0],
					"audio": w.sess.Ports[1]},
			},
		},
	}

	var reply interface{}
	if err := RPC(w.card.URL, "rtsp_client.query", params, &reply); err != nil {
		ret = false
	}

	rtspStatus := reply.(map[string]interface{})["transponds"].([]interface{})[0].(map[string]interface{})["status"]

	if rtspStatus != "Established" {
		ret = false
	}
	return
}

// Encode method
func (w *RTSPInWorker) Encode(sess *Session) error {

	settings := map[string]interface{}{
		"send_ip": sess.IP.String(),
		"video":   sess.Ports[0],
		"audio":   sess.Ports[1],
	}
	w.sess = *sess

	if err := w.set(w.workerID, settings); err != nil {
		return err
	}

	return nil
}

func (w *RTSPInWorker) set(id int, settings map[string]interface{}) error {
	w.card.lock.Lock()

	defer w.card.lock.Unlock()

	for k := range settings {
		helperSetMap(w.rpc, 0, k, settings[k])
	}

	// hack: ["rtsp_url"] must be set
	if w.rpc["transponds"].([]interface{})[0].(map[string]interface{})["rtsp_url"].(string) == "" {
		return nil
	}

	reply := make(map[string]interface{})
	if err := RPC(w.card.URL, "rtsp_client.add", w.rpc, &reply); err != nil {
		return err
	}

	if reply["transponds"].([]interface{})[0].(map[string]interface{})["status"].(string) != "Established" {
		return errInputError
	}

	return nil
}
