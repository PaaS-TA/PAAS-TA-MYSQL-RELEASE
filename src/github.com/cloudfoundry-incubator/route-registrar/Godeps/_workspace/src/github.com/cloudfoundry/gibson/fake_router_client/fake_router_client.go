package fake_gibson

type URIPortRoutePair struct {
	port int
	uri  string
}

type FakeRouterClient struct {
	DidGreet         bool
	RegisteredRoutes []URIPortRoutePair
}

func NewFakeRouterClient() *FakeRouterClient {
	return &FakeRouterClient{}
}

func (r *FakeRouterClient) Greet() error {
	r.DidGreet = true
	return nil
}

func (r *FakeRouterClient) Register(port int, uri string) error {
	r.RegisteredRoutes = append(r.RegisteredRoutes, URIPortRoutePair{port, uri})
	return nil
}

func (r *FakeRouterClient) Unregister(port int, uri string) error {
	for index, uriPortPair := range r.RegisteredRoutes {
		if uriPortPair.port == port && uriPortPair.uri == uri {
			r.RegisteredRoutes = append(r.RegisteredRoutes[:index], r.RegisteredRoutes[index+1:]...)
			break
		}
	}

	return nil
}

func (r *FakeRouterClient) IsRegistered(port int, uri string) bool {
	for _, uriPortPair := range r.RegisteredRoutes {
		if uriPortPair.port == port && uriPortPair.uri == uri {
			return true
		}
	}

	return false
}

func (r *FakeRouterClient) Reset() {
	r.DidGreet = false
	r.RegisteredRoutes = []URIPortRoutePair{}
}
