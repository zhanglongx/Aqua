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

type reg struct {
}

type regInfo struct {
	slot int

	name string

	ip net.IP
}

func (r *reg) Register() (Workers, error) {

	// tempz
	var cards []regInfo = []regInfo{
		{0, "local_encoder", net.IPv4(192, 165, 56, 35)},
	}

	var retErr error = nil
	workers := make(Workers)

	// TODO: prevent re-register
	for _, found := range cards {
		var card driver.Card
		switch found.name {
		case "local_encoder":
			card = &driver.LocalE{}
		default:
			comm.Error.Printf("unknown card type %s", found.name)
			retErr = errors.New("unknown card")
			continue
		}

		if w, err := card.Open(found.slot, found.ip); err == nil {
			workers[found.slot] = w
		}
	}

	return workers, retErr
}
