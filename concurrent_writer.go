package bdk

import (
	"bufio"
	"context"
	"time"
)

type ConcurrentWriter struct {
	w      *bufio.Writer
	jobs   chan []byte
	done   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

func NewConcurrentWriter(w *bufio.Writer) *ConcurrentWriter {
	ctx, cancel := context.WithCancel(context.Background())
	return &ConcurrentWriter{
		w:      w,
		jobs:   make(chan []byte, 1000),
		done:   make(chan struct{}),
		ctx:    ctx,
		cancel: cancel,
	}
}
func (w ConcurrentWriter) Write(msg []byte) (n int, err error) {
	if IsCtxDone(w.ctx) {
		return 0, nil
	}
	bts := make([]byte, len(msg))
	copy(bts, msg)
	w.jobs <- bts
	return len(msg), nil
}
func (w ConcurrentWriter) Consume() {
	go func() {
		for {
			time.Sleep(time.Second)
			_ = w.w.Flush()
			if _, more := <-w.jobs; !more {
				return
			}
		}
	}()
	go func() {
		for msg := range w.jobs {
			for i := 0; i < 10; i++ {
				n, _ := w.w.Write(msg)
				if n == len(msg) {
					break
				}
				msg = msg[n:]
			}
		}
		<-w.done
	}()
}
func (w ConcurrentWriter) Close() error {
	w.cancel()
	close(w.jobs)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-w.done:
			break loop
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
	return w.w.Flush()
}
func (w ConcurrentWriter) Sync() error {
	return w.w.Flush()
}
