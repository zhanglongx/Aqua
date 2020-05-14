package driver

import (
	"fmt"
	"net"
	"sync"

	"github.com/zhanglongx/Aqua/comm"
)

const (
	// C9810Name C9810 encoder card, one encoder worker(two channel)
	C9810Name string = "C9810"
	// C9811Name C9811 encoder card
	C9811Name string = "C9811"
)

// C981X info struct
type C981X struct {
	sync.RWMutex
	CardName string                 // current card name
	Slot     int                    // slot id
	IP       net.IP                 // card ip address
	URL      string                 // json-rpc interface
	rpc      map[string]interface{} // parameter table
}

// C981XWorker C981X worker info struct
type C981XWorker struct {
	workerID int
	card     *C981X
}

// Open C981X return 2 workers
func (c *C981X) Open() ([]Worker, error) {
	args := map[string]interface{}{}
	c.rpc = make(map[string]interface{})
	if err := RPC(c.URL, "encoder.get", args, &c.rpc); err != nil {
		return nil, err
	}
	return []Worker{
		&C981XWorker{
			workerID: 0,
			card:     c,
		},
	}, nil
}

// Close ...
func (c *C981X) Close() error {
	return nil
}

// Control Control method via CtlCmd and workerID
func (w *C981XWorker) Control(c CtlCmd, arg interface{}) interface{} {
	card := w.card

	switch c {
	case CtlCmdStart:
		settings := map[string]interface{}{
			"ctrl":     1,
			"enc_ctrl": 1,
		}
		if err := card.set(w.workerID, settings); err != nil {
			return err
		}
	case CtlCmdStop:
		settings := map[string]interface{}{
			"ctrl":     0,
			"enc_ctrl": 0,
		}
		if err := card.set(w.workerID, settings); err != nil {
			return err
		}
	case CtlCmdName:
		return fmt.Sprintf("%s_%d_%d", card.CardName,
			card.Slot, w.workerID)
	case CtlCmdIP:
		return card.IP
	case CtlCmdWorkerID:
		return w.workerID
	case CtlCmdSetting: // note: parameters in C981X have no array, workerID should be 0
		if settings, ok := arg.(map[string]interface{}); ok {
			if err := card.set(0, settings); err != nil {
				return err
			}
		}
	}
	return nil
}

// Monitor .
func (w *C981XWorker) Monitor() bool {
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
func (w *C981XWorker) Encode(sess *Session) error {
	settings := map[string]interface{}{
		"ip_send_addr": sess.IP.String(),
		"ip_send_port": sess.Ports[0],
	}
	if err := w.card.set(0, settings); err != nil {
		comm.Error.Println(sess.IP.String())
		return err
	}
	return nil
}

func (c *C981X) set(id int, settings map[string]interface{}) error {
	c.Lock()
	defer c.Unlock()

	for k := range settings {
		helperSetMap(c.rpc, id, k, settings[k])
	}

	var ok string
	if err := RPC(c.URL, "encoder.set", c.rpc, &ok); err != nil {
		return err
	}
	return nil
}
