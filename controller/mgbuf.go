package controller

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
)


type MsgBufSender struct {
	ctx        context.Context
	ctrl       *Controller
	recivechan chan any
	closed     atomic.Bool
	mu         sync.RWMutex
	timeoutset bool
	canc context.CancelFunc
}

// message buffer and sender for loopopration like watchman refreshdb & configiniter
func NewBufSender(ctx context.Context, ctrl *Controller, buf int, timeout time.Duration) *MsgBufSender {
	bufmg := &MsgBufSender{
		ctx:        ctx,
		ctrl:       ctrl,
		recivechan: make(chan any, buf),
	}
	if timeout != 0 {
		bufmg.timeoutset = true
		bufmg.ctx, bufmg.canc = context.WithTimeout(ctx, timeout)
	}
	bufmg.timeoutset = true
	return bufmg
}

func (m *MsgBufSender) Start() error {
	if m.ctx == nil {
		m.ctx = context.Background()
	}
	var wg sync.WaitGroup

	var currentcounter int
	for val := range m.recivechan {
		if m.ctx.Err() != nil {
			m.close()
			return m.ctx.Err()
		}

		if _, ok := val.(uint16); ok {
			m.close()
			return nil
		}

		currentcounter++
		wg.Add(1)
		go func() {
			_, err := m.ctrl.SendMsgContext(m.ctx, val)
			defer wg.Done()
			if err != nil {
				if errors.Is(err, C.ErrClientRequestFail) && !m.closed.Load() {
					m.mu.RLock()
					select {
					case m.recivechan <- val:
					default:
					}
					m.mu.RUnlock()
				}
			}
		}()

		if currentcounter >= 10 {
			wg.Wait()
			currentcounter = 0
		}
	}

	return nil
}

func (m *MsgBufSender) Send(msg string, id int64) {
	if m.closed.Load() {
		return
	}


	msgg := &botapi.Msgcommon{
		Infocontext: &botapi.Infocontext{
			ChatId:  id,
			User_id: id,
		},
		Text: msg,
	}

	m.mu.RLock()
	select {
	case m.recivechan <- msgg:
	default:
		m.mu.RUnlock()
		go m.ctrl.SendMsgContext(m.ctx, msgg)
		return
	}
	m.mu.RUnlock()
}

func (m *MsgBufSender) Over() {
	if m.closed.Load() {
		return
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.recivechan <- uint16(0)
}

// do not use this, only internal
func (m *MsgBufSender) close() error {
	if !m.closed.Swap(true) {
		m.mu.Lock()
		defer m.mu.Unlock()
		close(m.recivechan)
	}
	if m.canc != nil {
		m.canc()
	}
	return nil
}

//forceclose
func (m *MsgBufSender) Close() error {
	if m.timeoutset {
		//will be close automatically
		return nil		
	}
	return m.close()
}