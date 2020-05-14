// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package manager

import (
	"errors"
	"fmt"
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

	url string
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

	// tempz
	// cards = append(cards, regInfo{32, "local_encoder", net.IPv4(192, 165, 53, 35), ""})
	// cards = append(cards, regInfo{33, "local_decoder", net.IPv4(192, 165, 53, 35), ""})
	// cards = append(cards, regInfo{5, "C9810", net.IPv4(10, 1, 41, 180), "http://10.1.41.180/goform/form_data"})
	// cards = append(cards, regInfo{8, "C9811", net.IPv4(10, 1, 41, 180), "http://10.1.41.180/goform/form_data"})
	cards = append(cards, regInfo{6, "C9820Dec", net.IPv4(10, 1, 41, 179), "http://10.1.41.179/goform/form_data"})
	cards = append(cards, regInfo{6, "C9820Enc", net.IPv4(10, 1, 41, 179), "http://10.1.41.179/goform/form_data"})
	cards = append(cards, regInfo{7, "9550Av3Dec", net.IPv4(10, 1, 41, 157), "http://10.1.41.157/goform/form_data"})
	cards = append(cards, regInfo{7, "9550Av3Enc", net.IPv4(10, 1, 41, 157), "http://10.1.41.157/goform/form_data"})

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
		// local encoder
		case driver.LocalEncoderName:
			card = &driver.LocalE{Slot: found.slot,
				IP: found.ip,
			}
		// local decoder
		case driver.LocalDecoderName:
			card = &driver.LocalD{Slot: found.slot,
				IP: found.ip,
			}
		// C9830: rtsp + C9830(transcoder)
		case "C9830":
			card9830 := &driver.C9830{Slot: found.slot,
				IP:  found.ip,
				URL: found.url,
			}

			cardRTSP := &driver.RTSPIn{Slot: 255,
				IP:  comm.AppCfg.TransitSvr,
				URL: fmt.Sprintf("http://%s/goform/form_data", comm.AppCfg.TransitSvr),
			}

			card = &driver.TCBin{Card9830: card9830,
				CardRTSP: cardRTSP,
			}
		// C9820Enc
		case driver.C9820EncName:
			card = &driver.C9820Enc{C9820: driver.C9820{Slot: found.slot, IP: found.ip, URL: found.url}}
		// C9820Dec
		case driver.C9820DecName:
			card = &driver.C9820Dec{C9820: driver.C9820{Slot: found.slot, IP: found.ip, URL: found.url}}
		// F1000: test fake encode card
		case driver.F1000Name:
			card = &driver.F1000{Slot: found.slot, IP: found.ip}
		// F2000: test fake decode card
		case driver.F2000Name:
			card = &driver.F2000{Slot: found.slot, IP: found.ip}
		// 9550Av3: encode channel
		case driver.D9550Av3EncName:
			card = &driver.D9550Av3Enc{Slot: found.slot, IP: found.ip, URL: found.url}
		// 9550Av3: decode channel
		case driver.D9550Av3DecName:
			card = &driver.D9550Av3Dec{Slot: found.slot, IP: found.ip, URL: found.url}
		// C9810: encode card
		case driver.C9810Name, driver.C9811Name:
			card = &driver.C981X{CardName: found.name, Slot: found.slot, IP: found.ip, URL: found.url}
		default:
			comm.Error.Printf("Unknown card: %s", found.name)
			continue
		}

		comm.Info.Printf("Registering card %s in slot %d %v",
			found.name, found.slot, found.ip)

		if alloced[found.slot] == true {
			comm.Error.Printf("Slot %d already registered", found.slot)
			continue
		}

		if w, err := card.Open(); err == nil {
			*ws = append(*ws, w...)
			alloced[found.slot] = true
		} else {
			comm.Error.Printf("Open card %s failed: %s", found.name, err)
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

	args := map[string]interface{}{"cards": [0]int{}}

	var reply map[string]interface{}
	if err := driver.RPC(driver.TransURL,
		"register_server.query", args, &reply); err != nil {
		return nil, err
	}

	var result []regInfo

	var v interface{}
	for _, v = range reply["cards"].([]interface{}) {
		name := v.(map[string]interface{})["name"].(string)
		slot := int(v.(map[string]interface{})["slot"].(float64))

		cpus := v.(map[string]interface{})["cpus"].([]interface{})[0]
		ip := net.ParseIP(cpus.(map[string]interface{})["ip"].(string))

		url := v.(map[string]interface{})["url"].(string)

		result = append(result, regInfo{slot: slot, name: name, ip: ip, url: url})

		comm.Info.Printf("Found %s ip: %s slot: %d", name, ip, slot)
	}

	return result, nil
}
