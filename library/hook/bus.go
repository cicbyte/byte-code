package hook

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
)

type Event struct {
	Name    string
	Payload g.Map
}

type Handler func(ctx context.Context, event Event) error

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

var defaultBus = &Bus{
	handlers: make(map[string][]Handler),
}

func Default() *Bus {
	return defaultBus
}

func (b *Bus) Subscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

func (b *Bus) Emit(ctx context.Context, event Event) {
	b.mu.RLock()
	handlers, ok := b.handlers[event.Name]
	b.mu.RUnlock()

	if !ok {
		return
	}

	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			g.Log().Warningf(ctx, "Hook handler error for event %s: %v", event.Name, err)
		}
	}
}

func Subscribe(eventName string, handler Handler) {
	defaultBus.Subscribe(eventName, handler)
}

func Emit(ctx context.Context, eventName string, payload g.Map) {
	defaultBus.Emit(ctx, Event{Name: eventName, Payload: payload})
}
