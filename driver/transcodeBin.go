// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package driver

// TranscoderBinName is the sub-card's name
const TranscoderBinName string = "TransCoder"

// TCBin is the main struct for the Bin
type TCBin struct {
	Card9830 Card
	CardRTSP Card

	c9830Ws []Worker
	rtspWs  []Worker

	svr *PipeSvr
}

// TCBinWorker is the main struct for sub-card's
// Worker
type TCBinWorker struct {
	workerID int

	bin *TCBin
}

// Open method
func (b *TCBin) Open() ([]Worker, error) {
	b.svr = Pipes[PipeRTSPIN]

	var err error
	if b.c9830Ws, err = b.Card9830.Open(); err != nil {
		return nil, err
	}

	if b.rtspWs, err = b.CardRTSP.Open(); err != nil {
		return nil, err
	}

	ws := []Worker{}
	for id := range b.c9830Ws {
		rtspWorker := b.rtspWs[id]
		if err := b.svr.AllocPush(id+1, rtspWorker); err != nil {
			return nil, err
		}

		C9830Worker := b.c9830Ws[id]
		if err := b.svr.AllocPull(id+1, C9830Worker); err != nil {
			return nil, err
		}

		ws = append(ws, &TCBinWorker{workerID: id, bin: b})
	}

	return ws, nil
}

// Close method
func (b *TCBin) Close() error {

	for id := range b.c9830Ws {
		if err := b.svr.FreePush(id); err != nil {
			return err
		}

		C9830Worker := b.c9830Ws[id]
		if err := b.svr.FreePull(id, C9830Worker); err != nil {
			return err
		}
	}

	b.Card9830.Close()
	b.CardRTSP.Close()

	return nil
}

// Control method
func (w *TCBinWorker) Control(c CtlCmd, arg interface{}) interface{} {
	C9830Worker := w.bin.c9830Ws[w.workerID]
	rtspWorker := w.bin.rtspWs[w.workerID]

	switch c {
	case CtlCmdStart:
		if err := C9830Worker.Control(CtlCmdStart, nil); err != nil {
			return err
		}

		if err := rtspWorker.Control(CtlCmdStart, nil); err != nil {
			return err
		}

	case CtlCmdStop:
		if err := C9830Worker.Control(CtlCmdStop, nil); err != nil {
			return err
		}

		if err := rtspWorker.Control(CtlCmdStop, nil); err != nil {
			return err
		}

	case CtlCmdName:
		return C9830Worker.Control(CtlCmdName, nil)

	case CtlCmdIP:
		return C9830Worker.Control(CtlCmdName, nil)

	case CtlCmdWorkerID:
		return w.workerID

	case CtlCmdSetting:
		var settings map[string]interface{}
		var ok bool
		if settings, ok = arg.(map[string]interface{}); !ok {
			return errTypeError
		}

		if _, ok = settings["rtsp_url"]; ok {
			if err := rtspWorker.Control(CtlCmdSetting, settings); err != nil {
				return err
			}
		} else {
			return errKeyError
		}

	default:
	}
	return nil
}

// Monitor .
func (w *TCBinWorker) Monitor() bool {
	return true
}

// Encode method
func (w *TCBinWorker) Encode(sess *Session) error {

	C9830Worker := w.bin.c9830Ws[w.workerID]
	if err := SetEncodeSes(C9830Worker, sess); err != nil {
		return err
	}

	return nil
}
