// Package closer предоставляет механизм для регистрации и последовательного
// закрытия ресурсов (соединений, файлов и т.д.) при завершении работы приложения.
package closer

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Func представляет функцию завершения, принимающую контекст и возвращающую ошибку в случае неудачного завершения.
type Func func(ctx context.Context) error

// Closer управляет списком функций завершения и обеспечивает их вызов в LIFO-порядке.
type Closer struct {
	mu    sync.Mutex
	funcs []Func
}

// NewCloser создает новый экземпляр Closer.
func NewCloser() *Closer {
	return &Closer{}
}

// Add добавляет функцию завершения в стек Closer.
func (c *Closer) Add(f Func) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.funcs = append(c.funcs, f)
}

// Close вызывает все зарегистрированные функции завершения в обратном порядке.
// Если одна или несколько функций возвращают ошибку, ошибки собираются и возвращаются одной.
// Если контекст истекает до завершения всех функций, возвращается ошибка таймаута.
func (c *Closer) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var (
		msgs    = make([]string, 0, len(c.funcs))
		wg      sync.WaitGroup
		errorCh = make(chan error, len(c.funcs))
		done    = make(chan struct{})
	)

	// We finish in LIFO order
	for i := len(c.funcs) - 1; i >= 0; i-- {
		wg.Add(1)
		go func(f Func) {
			defer wg.Done()
			if err := f(ctx); err != nil {
				errorCh <- err
			}
		}(c.funcs[i])
	}

	go func() {
		wg.Wait()
		close(done)
		close(errorCh)
	}()

	select {
	case <-done:
		break
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %v", ctx.Err())
	}

	for err := range errorCh {
		msgs = append(msgs, fmt.Sprintf("[!] %v", err))
	}

	if len(msgs) > 0 {
		return fmt.Errorf(
			"shutdown completed with errors:\n%s",
			strings.Join(msgs, "\n"),
		)
	}

	return nil
}
