package driver

import (
	"fmt"
	"net"
	"sync"
)

const (
	// C9820DecName C9820 decoder card name
	C9820DecName = "C9820Dec"
	// C9820EncName C9820 encoder card name
	C9820EncName = "C9820Enc"
)

// C9820 comm data
type C9820 struct {
	Slot int
	IP   net.IP
	URL  string
}

type c9820Rpc struct {
	sync.RWMutex
	rpc map[string]interface{}
}

var rpcObj c9820Rpc

// C9820Dec decoder channel card
type C9820Dec struct {
	C9820
	commRPC *c9820Rpc
}

// C9820Enc encoder channel card
type C9820Enc struct {
	C9820
	commRPC *c9820Rpc
}

// C9820DecWorker worker info struct
type C9820DecWorker struct {
	workerID int
	card     *C9820Dec
}

// C9820EncWorker worker info struct
type C9820EncWorker struct {
	workerID int
	card     *C9820Enc
}

// Open encoder channel Open method
func (c *C9820Enc) Open() ([]Worker, error) {
	c.commRPC = &rpcObj
	if err := commInit(c.commRPC, c.URL, "mosaic_encoder.get"); err != nil {
		return nil, err
	}
	ws := []Worker{
		&C9820EncWorker{
			workerID: 9,
			card:     c,
		},
	}
	return ws, nil
}

// Open decoder channel Open method
func (c *C9820Dec) Open() ([]Worker, error) {
	c.commRPC = &rpcObj
	if err := commInit(c.commRPC, c.URL, "mosaic_encoder.get"); err != nil {
		return nil, err
	}
	ws := []Worker{}
	for i := 0; i < 9; i++ {
		ws = append(ws, &C9820DecWorker{
			workerID: i,
			card:     c,
		})
	}
	return ws, nil
}

func commInit(rpcTmp *c9820Rpc, URL, method string) error {
	if rpcTmp.rpc != nil {
		fmt.Println("inited")
		return nil
	}
	fmt.Println("initing")
	rpcTmp.Lock()
	defer rpcTmp.Unlock()
	args := map[string]interface{}{}
	rpcTmp.rpc = make(map[string]interface{})
	if err := RPC(URL, method, args, &rpcTmp.rpc); err != nil {
		return err
	}
	for i := 0; i < 9; i++ {
		helperSetMap(rpcTmp.rpc, i, "recv_cast_mode", 0)
	}
	helperSetMap(rpcTmp.rpc, 0, "mosic_mode", 3)

	var ok string
	if err := RPC(URL, "mosaic_encoder.set", rpcTmp.rpc, &ok); err != nil {
		return err
	}
	return nil
}

//Close ..
func (c *C9820Dec) Close() error {
	return nil
}

//Close ..
func (c *C9820Enc) Close() error {
	return nil
}

// Control ..
func (w *C9820EncWorker) Control(c CtlCmd, arg interface{}) interface{} {
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
		return fmt.Sprintf("%s_%d_%d", C9820EncName, card.Slot, w.workerID)
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

// Control ..
func (w *C9820DecWorker) Control(c CtlCmd, arg interface{}) interface{} {
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
		return fmt.Sprintf("%s_%d_%d", C9820DecName, card.Slot, w.workerID)
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
func (w *C9820EncWorker) Monitor() bool {
	params := map[string]interface{}{
		"send": map[string]interface{}{
			"bitrate": 0,
		},
	}
	var reply interface{}
	if err := RPC(w.card.URL, "mosaic_encoder.get", params, &reply); err != nil {
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

// Monitor .
func (w *C9820DecWorker) Monitor() bool {
	return true
}

// Encode mosaic encoder, note there is only one encoder
func (w *C9820EncWorker) Encode(sess *Session) error {
	settings := map[string]interface{}{
		"send_ip":   sess.IP.String(),
		"send_port": sess.Ports[0],
	}
	if err := w.card.set(0, settings); err != nil {
		return err
	}
	return nil
}

// Decode ..
func (w *C9820DecWorker) Decode(sess *Session) error {
	settings := map[string]interface{}{
		"vid_port": sess.Ports[0],
	}
	if err := w.card.set(w.workerID, settings); err != nil {
		return err
	}
	return nil
}

func (c *C9820Dec) set(id int, settings map[string]interface{}) error {
	c.commRPC.Lock()

	defer c.commRPC.Unlock()

	for k := range settings {
		helperSetMap(c.commRPC.rpc, id, k, settings[k])
	}

	var ok string
	if err := RPC(c.URL, "mosaic_encoder.set", c.commRPC.rpc, &ok); err != nil {
		return err
	}

	return nil
}
func (c *C9820Enc) set(id int, settings map[string]interface{}) error {
	c.commRPC.Lock()

	defer c.commRPC.Unlock()

	for k := range settings {
		helperSetMap(c.commRPC.rpc, id, k, settings[k])
	}

	var ok string
	if err := RPC(c.URL, "mosaic_encoder.set", c.commRPC.rpc, &ok); err != nil {
		return err
	}

	return nil
}
