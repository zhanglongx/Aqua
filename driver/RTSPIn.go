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
}

func newRPC() map[string]interface{} {

	// tempz ip
	return map[string]interface{}{
		"transponds": []map[string]interface{}{
			{
				"type":     "udp2udp",
				"rtsp_url": "",
				"recv_ip":  "10.1.41.152",
				"recv_port": map[string]interface{}{
					"video": 0,
					"audio": 0},
				"send_ip": "10.1.41.152",
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
			rpc:      newRPC(),
		},
		&RTSPInWorker{
			workerID: 1,
			card:     c,
			rpc:      newRPC(),
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

// Encode method
func (w *RTSPInWorker) Encode(sess *Session) error {

	settings := map[string]interface{}{
		"send_ip":   sess.IP.String(),
		"send_port": sess.Ports[0],
	}

	if err := w.set(w.workerID, settings); err != nil {
		return err
	}

	return nil
}

func (w *RTSPInWorker) set(id int, settings map[string]interface{}) error {
	w.card.lock.Lock()

	defer w.card.lock.Unlock()

	for k := range settings {
		helperSetMap(w.rpc, id, k, settings[k])
	}

	// hack: use ["rtsp_url"] to see if we can add
	if w.rpc["rtsp_url"] == "" {
		return nil
	}

	var reply map[string]interface{}
	var err error
	if reply, err = RPC(w.card.URL, "rtsp_client.add", w.rpc); err != nil {
		return err
	}

	w.rpc = reply

	return nil
}
