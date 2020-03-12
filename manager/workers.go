// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package manager

import (
	"errors"
	"net"

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
func (ws *Workers) register() error {

	// tempz
	var cards []regInfo = []regInfo{
		{0, "local_encoder", net.IPv4(192, 165, 56, 35)},
	}

	alloced := make(map[int]bool)

	for _, found := range cards {
		var card driver.Card
		switch found.name {
		case driver.LocalEncoderName:
			card = &driver.LocalE{}
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
