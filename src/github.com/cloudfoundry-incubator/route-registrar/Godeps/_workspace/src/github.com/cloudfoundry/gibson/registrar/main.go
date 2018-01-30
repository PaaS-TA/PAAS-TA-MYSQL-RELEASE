package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/gibson"
	"github.com/cloudfoundry/yagnats"
)

var ip = flag.String("ip", "", "IP address of the machine to route to")

var routes = flag.String("routes", "", "routes to register, in the form of port:uri,port:uri,port:uri")

var natsAddresses = flag.String("natsAddresses", "", "comma-separated list of NATS cluster member IP:ports")
var natsUsername = flag.String("natsUsername", "", "authentication user for connecting to NATS")
var natsPassword = flag.String("natsPassword", "", "authentication password for connecting to NATS")

func main() {
	flag.Parse()

	nats := yagnats.NewClient()

	natsMembers := []yagnats.ConnectionProvider{}

	if *natsAddresses == "" {
		log.Fatalln("must specify at least one nats address (-natsAddresses=1.2.3.4:5678)")
	}

	if *ip == "" {
		log.Fatalln("must specify IP to route to (-ip=X)")
	}

	for _, addr := range strings.Split(*natsAddresses, ",") {
		log.Println("configuring nats server:", addr)
		natsMembers = append(natsMembers, &yagnats.ConnectionInfo{
			Addr:     addr,
			Username: *natsUsername,
			Password: *natsPassword,
		})
	}

	if len(natsMembers) == 0 {
		log.Fatalln("must specify at least one nats address")
	}

	natsInfo := &yagnats.ConnectionCluster{natsMembers}

	for {
		err := nats.Connect(natsInfo)
		if err == nil {
			break
		}

		log.Println("failed to connect to NATS:", err)
		time.Sleep(1 * time.Second)
	}

	client := gibson.NewCFRouterClient(*ip, nats)

	client.Greet()

	for _, route := range strings.Split(*routes, ",") {
		routePair := strings.Split(route, ":")
		if len(routePair) != 2 {
			log.Fatalln("invalid route configuration:", *routes)
		}

		port, err := strconv.Atoi(routePair[0])
		if err != nil {
			log.Fatalln("invalid route port:", err)
		}

		client.Register(port, routePair[1])
	}

	select {}
}
