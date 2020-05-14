package driver

import (
	"fmt"
	"net"
	"sync"
)

// D9550Av3DecName .
const D9550Av3DecName string = "9550Av3Dec"

// D9550Av3Dec .
type D9550Av3Dec struct {
	sync.RWMutex
	Slot int
	IP   net.IP
	URL  string
	rpc  map[string]interface{}
	// 	rpcDecoder map[string]interface{}
	// 	rpcReceive map[string]interface{}
}

// D9550Av3DecWorker .
type D9550Av3DecWorker struct {
	workerID int
	card     *D9550Av3Dec
}

// Open .
func (d *D9550Av3Dec) Open() ([]Worker, error) {
	args := map[string]interface{}{}
	rpcDecoder := make(map[string]interface{})
	rpcReceive := make(map[string]interface{})
	d.rpc = make(map[string]interface{})
	d.rpc["decoder"] = rpcDecoder
	d.rpc["receive"] = rpcReceive
	if err := RPC(d.URL, "decoder.get", args, &rpcDecoder); err != nil {
		return nil, err
	}
	if err := RPC(d.URL, "receive.get", args, &rpcReceive); err != nil {
		return nil, err
	}
	return []Worker{
		&D9550Av3DecWorker{
			workerID: 0,
			card:     d,
		},
	}, nil
}

// Close .
func (d *D9550Av3Dec) Close() error {
	return nil
}

// Control .
func (w *D9550Av3DecWorker) Control(c CtlCmd, arg interface{}) interface{} {
	card := w.card
	switch c {
	case CtlCmdStart:
		settings := map[string]interface{}{
			"decoder": map[string]interface{}{
				".recv.ctrl":     1, // 1为ip接收
				".vdec.dec_ctrl": 1,
				".adec.ctrl":     1,
			},
			"receive": map[string]interface{}{
				".udp_recv.ctrl": 1,
				// ".udp_send.ctrl": 1, // uds发送开关
			},
		}
		if err := card.set(w.workerID, settings); err != nil {
			return err
		}
	case CtlCmdStop:
		settings := map[string]interface{}{
			"decoder": map[string]interface{}{
				".recv.ctrl":     0,
				".vdec.dec_ctrl": 0,
				".adec.ctrl":     0,
			},
			"receive": map[string]interface{}{
				".udp_recv.ctrl": 0,
				// ".udp_send.ctrl": 0, // uds发送开关
			},
		}
		if err := card.set(w.workerID, settings); err != nil {
			return err
		}
	case CtlCmdName:
		return fmt.Sprintf("%s_%d_%d", D9550Av3DecName,
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
	}
	return nil
}

// Monitor .
func (w *D9550Av3DecWorker) Monitor() bool {
	return true
}

// Decode .
func (w *D9550Av3DecWorker) Decode(sess *Session) error {
	settings := map[string]interface{}{
		"receive": map[string]interface{}{
			".udp_recv.recv_port": sess.Ports[0],
		},
	}
	if err := w.card.set(w.workerID, settings); err != nil {
		return err
	}
	return nil
}

func (d *D9550Av3Dec) set(id int, settings map[string]interface{}) error {
	d.Lock()
	defer d.Unlock()

	var reply string
	if _, ok := settings["decoder"]; ok {
		for k, v := range settings["decoder"].(map[string]interface{}) {
			if chk := helperSetParam(d.rpc["decoder"], "", k, v); chk == nil {
				fmt.Printf("param `%s` does not exist\n", k)
			}
		}
		if err := RPC(d.URL, "decoder.set", d.rpc["decoder"], &reply); err != nil {
			return err
		}
	}

	if _, ok := settings["receive"]; ok {
		for k, v := range settings["receive"].(map[string]interface{}) {
			if chk := helperSetParam(d.rpc["receive"], "", k, v); chk == nil {
				fmt.Printf("param `%s` does not exist\n", k)
			}
		}

		if err := RPC(d.URL, "receive.set", d.rpc["receive"], &reply); err != nil {
			return err
		}
	}
	return nil
}
