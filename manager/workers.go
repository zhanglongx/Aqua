// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package manager

import (
	"net"

	"github.com/zhanglongx/Aqua/comm"
	"github.com/zhanglongx/Aqua/driver"
)

// Workers store all workers registered by cards
type Workers []driver.Worker

type reg struct {
}

type regInfo struct {
	slot int

	name string

	ip net.IP
}

// register accept sub-card's register
func (ws *Workers) register() error {

	// tempz
	var cards []regInfo = []regInfo{
		{0, "local_encoder", net.IPv4(192, 165, 56, 35)},
	}

	alloced := make(map[int]bool)

	// TODO: prevent re-register
	for _, found := range cards {
		var card driver.Card
		switch found.name {
		case "local_encoder":
			card = &driver.LocalE{}
			comm.Info.Printf("registering card %s in slot %d %v",
				found.name, found.slot, found.ip)
		default:
			comm.Error.Printf("unknown card type %s", found.name)
			continue
		}

		if alloced[found.slot] == true {
			comm.Error.Print("slot already registered")
			continue
		}

		if w, err := card.Open(found.slot, found.ip); err == nil {
			*ws = append(*ws, w...)
			alloced[found.slot] = true
		} else {
			comm.Error.Printf("open card %s failed", found.name)
		}
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
