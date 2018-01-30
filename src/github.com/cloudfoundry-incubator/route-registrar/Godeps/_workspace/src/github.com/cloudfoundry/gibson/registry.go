package gibson

import (
	"sync"
)

type Registry struct {
	routes map[int]map[string]bool
	lock   sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		routes: make(map[int]map[string]bool),
	}
}

func (r *Registry) Register(port int, uri string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.routes[port] == nil {
		r.routes[port] = make(map[string]bool)
	}

	r.routes[port][uri] = true
}

func (r *Registry) Unregister(port int, uri string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.routes[port], uri)
	if len(r.routes[port]) == 0 {
		delete(r.routes, port)
	}
}

func (r *Registry) InParallel(callback func(int, []string)) (count int) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for port, urisMap := range r.routes {
		urisArray := make([]string, len(urisMap))
		index := 0
		for uri, _ := range urisMap {
			urisArray[index] = uri
			index += 1
			count += 1
		}

		go callback(port, urisArray)
	}

	return
}
