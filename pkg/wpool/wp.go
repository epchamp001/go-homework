package wpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Job struct {
	Ctx    context.Context
	Do     func(context.Context) (any, error)
	Result chan Response
}

type Response struct {
	Val any
	Err error
}

type Pool struct {
	mu     sync.Mutex
	jobs   chan Job
	wg     sync.WaitGroup
	cancel context.CancelFunc
	size   int
}

func NewWorkerPool(size, queueLen int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		jobs:   make(chan Job, queueLen),
		cancel: cancel,
	}
	p.resizeLocked(ctx, size)
	return p
}

func (p *Pool) Submit(j Job) {
	p.jobs <- j
}

func (p *Pool) Resize(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.resizeLocked(context.Background(), n)
}

func (p *Pool) Stop() {
	p.cancel()
	p.wg.Wait()
}

var errPoison = errors.New("stop") // пилюля для остановки лишних воркеров

func (p *Pool) resizeLocked(ctx context.Context, n int) {
	delta := n - p.size
	switch {
	case delta > 0:
		for i := 0; i < delta; i++ {
			p.wg.Add(1)
			go p.worker(ctx)
		}
	case delta < 0:
		for i := 0; i < -delta; i++ {
			p.jobs <- Job{
				Ctx: ctx,
				Do: func(context.Context) (any, error) {
					return nil, errPoison
				},
				Result: make(chan Response, 1),
			}
		}
	}
	p.size = n
}

func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job := <-p.jobs:
			// если клиент уже отменил запрос
			if err := job.Ctx.Err(); err != nil {
				job.Result <- Response{Err: err}
				continue
			}

			func() {
				// ловим панику, чтобы воркер не умер
				defer func() {
					if r := recover(); r != nil {
						job.Result <- Response{Err: fmt.Errorf("panic: %v", r)}
					}
				}()

				val, err := job.Do(job.Ctx)
				// пилюля: просто закрываем воркер
				if errors.Is(err, errPoison) {
					return
				}
				job.Result <- Response{Val: val, Err: err}
			}()
		}
	}
}
