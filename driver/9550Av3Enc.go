package driver

import (
	"fmt"
	"net"
	"sync"
)

// D9550Av3EncName .
const D9550Av3EncName string = "9550Av3Enc"

// D9550Av3Enc .
type D9550Av3Enc struct {
	sync.RWMutex
	Slot int
	IP   net.IP
	URL  string
	rpc  map[string]interface{}
}

// D9550Av3EncWorker .
type D9550Av3EncWorker struct {
	workerID int
	card     *D9550Av3Enc
}

// Open .
func (d *D9550Av3Enc) Open() ([]Worker, error) {
	args := map[string]interface{}{}
	d.rpc = make(map[string]interface{})
	if err := RPC(d.URL, "encoder.get", args, &d.rpc); err != nil {
		return nil, err
	}
	return []Worker{
		&D9550Av3EncWorker{
			workerID: 0,
			card:     d,
		},
	}, nil
}

// Close .
func (d *D9550Av3Enc) Close() error {
	return nil
}

// Control .
func (w *D9550Av3EncWorker) Control(c CtlCmd, arg interface{}) interface{} {
	card := w.card
	switch c {
	case CtlCmdStart:
		settings := map[string]interface{}{
			".venc.enc_ctrl": 1,
			".send.ctrl":     1,
			".aenc.ctrl":     1,
		}
		if err := card.set(w.workerID, settings); err != nil {
			return err
		}
	case CtlCmdStop:
		settings := map[string]interface{}{
			".venc.enc_ctrl": 0,
			".send.ctrl":     0,
			".aenc.ctrl":     0,
		}
		if err := card.set(w.workerID, settings); err != nil {
			return err
		}
	case CtlCmdName:
		return fmt.Sprintf("%s_%d_%d", D9550Av3EncName,
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
func (w *D9550Av3EncWorker) Monitor() bool {
	params := map[string]interface{}{
		"send": map[string]interface{}{
			"bitrate": 0,
		},
	}
	var reply interface{}
	if err := RPC(w.card.URL, "encoder.get", params, &reply); err != nil {
		return false
	}
	r, ok := reply.(map[string]interface{})
	if !ok {
		return false
	}
	s, ok := r["send"].(map[string]interface{})
	if !ok {
		return false
	}
	b, ok := s["bitrate"].(float64)
	if !ok {
		return false
	}
	if b <= 0 {
		return false
	}
	return true
}

// Encode .
func (w *D9550Av3EncWorker) Encode(sess *Session) error {
	settings := map[string]interface{}{
		".send.ip_send_addr": sess.IP.String(),
		".send.ip_send_port": sess.Ports[0],
	}
	if err := w.card.set(w.workerID, settings); err != nil {
		return err
	}
	return nil
}

func (d *D9550Av3Enc) set(id int, settings map[string]interface{}) error {
	d.Lock()
	defer d.Unlock()

	for k, v := range settings {
		// when return is nil, `k` cannot be found in `m`
		if chk := helperSetParam(d.rpc, "", k, v); chk == nil {
			fmt.Printf("param `%s` does not exist\n", k)
		}
	}

	var ok string
	if err := RPC(d.URL, "encoder.set", d.rpc, &ok); err != nil {
		return err
	}
	return nil
}
