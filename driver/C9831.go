package driver

import (
	"fmt"
	"net"
	"sync"
)

// C9831SmartVideoName .
const C9831SmartVideoName string = "C9831"

// C9831 .
type C9831 struct {
	sync.RWMutex
	Slot int
	IP   net.IP
	URL  string
	rpc  map[string]interface{}
}

// C9831Worker .
type C9831Worker struct {
	workerID int
	card     *C9831
}

// Open .
func (c *C9831) Open() ([]Worker, error) {
	args := map[string]interface{}{}
	c.rpc = make(map[string]interface{})
	if err := RPC(c.URL, "smartvideo.get", args, &c.rpc); err != nil {
		return nil, err
	}

	for i := 0; i < 2; i++ {
		helperSetMap(c.rpc, i, "recv_cast_mode", 0)
	}

	var ok string
	if err := RPC(c.URL, "smartvideo.set", c.rpc, &ok); err != nil {
		return nil, err
	}
	return []Worker{}, nil
}

// Close .
func (c *C9831) Close() error {
	return nil
}

// Control .
func (w *C9831Worker) Control(c CtlCmd, arg interface{}) interface{} {
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

// Encode .
func (w *C9831Worker) Encode(sess *Session) error {
	settings := map[string]interface{}{
		"send_ip":   sess.IP.String(),
		"send_port": sess.Ports[0],
	}
	if err := w.card.set(w.workerID, settings); err != nil {
		return err
	}

	return nil
}

// Decode .
func (w *C9831Worker) Decode(sess *Session) error {
	settings := map[string]interface{}{
		"vid_port": sess.Ports[0],
	}
	if err := w.card.set(w.workerID, settings); err != nil {
		return err
	}

	return nil
}

func (c *C9831) set(id int, settings map[string]interface{}) error {
	c.Lock()

	defer c.Unlock()

	for k := range settings {
		helperSetMap(c.rpc, id, k, settings[k])
	}

	var ok string
	if err := RPC(c.URL, "smartvideo.set", c.rpc, &ok); err != nil {
		return err
	}

	return nil
}
