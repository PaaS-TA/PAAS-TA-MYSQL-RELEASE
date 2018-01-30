package gibson

import (
	"fmt"

	. "launchpad.net/gocheck"
)

type RSuite struct {
	registry *Registry
}

func init() {
	Suite(&RSuite{})
}

func (s *RSuite) SetUpTest(c *C) {
	s.registry = NewRegistry()
}

func (s *RSuite) AssertRegistryContains(c *C, ports []int, uris []string) {
	expectedPortUris := map[string]bool{}

	for i, _ := range ports {
		expectedPortUris[fmt.Sprintf("%d:%s", ports[i], uris[i])] = true
	}

	portUriInRegistry := make(chan string)
	count := s.registry.InParallel(func(port int, uris []string) {
		for _, uri := range uris {
			portUriInRegistry <- fmt.Sprintf("%d:%s", port, uri)
		}
	})

	c.Assert(count, Equals, len(ports))

	for _ = range ports {
		c.Assert(expectedPortUris[<-portUriInRegistry], Equals, true)
	}
}

func (s *RSuite) TestRegistryCRUD(c *C) {
	s.registry.Register(123, "foo.uri")
	s.registry.Register(123, "bar.uri")
	s.registry.Register(427, "tk")

	s.AssertRegistryContains(c, []int{123, 123, 427}, []string{"foo.uri", "bar.uri", "tk"})

	s.registry.Unregister(123, "foo.uri")

	s.AssertRegistryContains(c, []int{123, 427}, []string{"bar.uri", "tk"})

	s.registry.Unregister(123, "bar.uri")

	s.AssertRegistryContains(c, []int{427}, []string{"tk"})
}

func (s *RSuite) TestRegistryEdgeCases(c *C) {
	s.registry.Register(123, "foo.uri")
	s.registry.Unregister(123, "bar.uri")
	s.registry.Register(456, "foo.uri")

	s.AssertRegistryContains(c, []int{123, 456}, []string{"foo.uri", "foo.uri"})
}
