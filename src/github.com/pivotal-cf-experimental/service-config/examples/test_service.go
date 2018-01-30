package main

import (
	"github.com/pivotal-cf-experimental/service-config"

	"flag"
	"fmt"
	"os"
)

type ShipConfig struct {
	Name   string
	ID     int
	Crew   Crew
	Active bool
}

type Crew struct {
	Officers   []Officer
	Passengers []Passenger
}

type Officer struct {
	Name string
	Role string
}

type Passenger struct {
	Name  string
	Title string
}

func main() {
	serviceConfig := service_config.New()

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	serviceConfig.AddDefaults(ShipConfig{
		Active: true,
	})

	serviceConfig.AddFlags(flags)
	flags.Parse(os.Args[1:])

	var config ShipConfig
	err := serviceConfig.Read(&config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Config: %#v\n", config)
}
