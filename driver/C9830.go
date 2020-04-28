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

// C9830TranscoderName is the sub-card's name
const C9830TranscoderName string = "C9830"

// C9830 is the main struct for sub-card
type C9830 struct {
	lock sync.RWMutex

	// Card Slot
	Slot int

	// Card IP
	IP net.IP

	URL string

	rpc map[string]interface{}
}

// C9830Worker is the main struct for sub-card's
// Worker
type C9830Worker struct {
	workerID int

	card *C9830
}

// Open method
func (c *C9830) Open() ([]Worker, error) {
	args := map[string]interface{}{}

	c.rpc = make(map[string]interface{})
	if err := RPC(c.URL, "transcoder.get", args, &c.rpc); err != nil {
		return nil, err
	}

	// set to default
	for i := 0; i < 2; i++ {
		helperSetMap(c.rpc, i, "recv_cast_mode", 0)
	}

	var ok string
	if err := RPC(c.URL, "transcoder.set", c.rpc, &ok); err != nil {
		return nil, err
	}

	return []Worker{
		&C9830Worker{
			workerID: 0,
			card:     c,
		},
		&C9830Worker{
			workerID: 1,
			card:     c,
		},
	}, nil
}

// Close method
func (c *C9830) Close() error {
	return nil
}

// Control method
func (w *C9830Worker) Control(c CtlCmd, arg interface{}) interface{} {
	card := w.card

	switch c {
	case CtlCmdStart:
		settings := map[string]interface{}{
			"ctrl": 1,
		}
		if err := card.set(w.workerID, settings); err != nil {
			return err
		}

	case CtlCmdStop:
		settings := map[string]interface{}{
			"ctrl": 0,
		}
		if err := card.set(w.workerID, settings); err != nil {
			return err
		}

	case CtlCmdName:
		return fmt.Sprintf("%s_%d_%d", C9830TranscoderName,
			card.Slot, w.workerID)

	case CtlCmdIP:
		return card.IP

	case CtlCmdWorkerID:
		return w.workerID

	case CtlCmdSetting:
		if settings, ok := arg.(map[string]interface{}); ok {
			if err := card.set(w.workerID, settings); err != nil {
				return err
			}
		}

	default:
	}
	return nil
}

// Encode method
func (w *C9830Worker) Encode(sess *Session) error {
	settings := map[string]interface{}{
		"send_ip":   sess.IP.String(),
		"send_port": sess.Ports[0],
	}
	if err := w.card.set(w.workerID, settings); err != nil {
		return err
	}

	return nil
}

// Decode method
func (w *C9830Worker) Decode(sess *Session) error {
	settings := map[string]interface{}{
		"vid_port": sess.Ports[0],
	}
	if err := w.card.set(w.workerID, settings); err != nil {
		return err
	}

	return nil
}

func (c *C9830) set(id int, settings map[string]interface{}) error {
	c.lock.Lock()

	defer c.lock.Unlock()

	for k := range settings {
		helperSetMap(c.rpc, id, k, settings[k])
	}

	var ok string
	if err := RPC(c.URL, "transcoder.set", c.rpc, &ok); err != nil {
		return err
	}

	return nil
}
