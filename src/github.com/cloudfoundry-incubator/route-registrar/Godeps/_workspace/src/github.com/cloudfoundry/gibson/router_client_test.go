package gibson

import (
	"time"

	"github.com/cloudfoundry/yagnats"
	"github.com/cloudfoundry/yagnats/fakeyagnats"
	. "launchpad.net/gocheck"
)

type RCSuite struct{}

func init() {
	Suite(&RCSuite{})
}

func (s *RCSuite) TestRouterClientRegistering(c *C) {
	mbus := fakeyagnats.New()

	routerClient := NewCFRouterClient("1.2.3.4", mbus)

	routerClient.Register(123, "abc")

	registrations := mbus.PublishedMessages["router.register"]

	c.Assert(len(registrations), Not(Equals), 0)
	c.Assert(string(registrations[0].Payload), Equals, `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
}

func (s *RCSuite) TestRouterClientUnregistering(c *C) {
	mbus := fakeyagnats.New()

	routerClient := NewCFRouterClient("1.2.3.4", mbus)

	routerClient.Unregister(123, "abc")

	unregistrations := mbus.PublishedMessages["router.unregister"]

	c.Assert(len(unregistrations), Not(Equals), 0)
	c.Assert(string(unregistrations[0].Payload), Equals, `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
}

func (s *RCSuite) TestRouterClientRouterStartHandling(c *C) {
	mbus := fakeyagnats.New()

	routerClient := NewCFRouterClient("1.2.3.4", mbus)

	err := routerClient.Greet()
	c.Assert(err, IsNil)

	startCallback := mbus.Subscriptions["router.start"][0]
	startCallback.Callback(&yagnats.Message{
		Payload: []byte(`{"minimumRegisterIntervalInSeconds":1}`),
	})

	routerClient.Register(123, "abc")

	c.Assert(len(mbus.PublishedMessages["router.register"]), Equals, 1)

	time.Sleep(600 * time.Millisecond)

	c.Assert(len(mbus.PublishedMessages["router.register"]), Equals, 1)

	time.Sleep(600 * time.Millisecond)

	c.Assert(len(mbus.PublishedMessages["router.register"]), Equals, 2)

	time.Sleep(600 * time.Millisecond)

	c.Assert(len(mbus.PublishedMessages["router.register"]), Equals, 2)

	time.Sleep(600 * time.Millisecond)

	c.Assert(len(mbus.PublishedMessages["router.register"]), Equals, 3)
}

func (s *RCSuite) TestRouterClientGreeting(c *C) {
	mbus := fakeyagnats.New()

	routerClient := NewCFRouterClient("1.2.3.4", mbus)

	routerClient.Register(123, "abc")

	err := routerClient.Greet()
	c.Assert(err, IsNil)

	greetMsg := mbus.PublishedMessages["router.greet"][0]

	greetCallback := mbus.Subscriptions[greetMsg.ReplyTo][0]
	greetCallback.Callback(&yagnats.Message{
		Payload: []byte(`{"minimumRegisterIntervalInSeconds":1}`),
	})

	c.Assert(len(mbus.PublishedMessages["router.register"]), Equals, 1)

	time.Sleep(600 * time.Millisecond)

	c.Assert(len(mbus.PublishedMessages["router.register"]), Equals, 1)

	time.Sleep(600 * time.Millisecond)

	c.Assert(len(mbus.PublishedMessages["router.register"]), Equals, 2)

	time.Sleep(600 * time.Millisecond)

	c.Assert(len(mbus.PublishedMessages["router.register"]), Equals, 2)

	time.Sleep(600 * time.Millisecond)

	c.Assert(len(mbus.PublishedMessages["router.register"]), Equals, 3)
}
