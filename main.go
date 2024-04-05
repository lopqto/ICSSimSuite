package main

import (
	"fmt"
	"os"
	"time"

	config "github.com/lopqto/icssimsuite/pkg/config"
	handler "github.com/lopqto/icssimsuite/pkg/handlers"

	"github.com/simonvetter/modbus"
	log "github.com/sirupsen/logrus"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <config.toml>\n", os.Args[0])
		os.Exit(1)
	}

	configFile := os.Args[1]

	var server *modbus.ModbusServer
	var err error
	var gh *handler.Handler

	// create the config object
	c := config.Config{}
	_, err = c.LoadConfig(configFile)
	if err != nil {
		log.Errorf("Error: %v", err)
		os.Exit(1)
	}

	// set the log level
	log.SetLevel(c.MapLogLevel(c.LogLevel))
	log.Debugf("Config: %v", c)

	// create the handler object
	gh = handler.NewHandler(&c)

	// create the server object
	server, err = modbus.NewServer(&modbus.ServerConfiguration{
		URL:        fmt.Sprintf("tcp://%s:%d", c.Host, c.Port),
		Timeout:    time.Duration(c.IdleTimeout) * time.Second,
		MaxClients: c.MaxClients,
	}, gh)
	if err != nil {
		fmt.Printf("failed to create server: %v\n", err)
		os.Exit(1)
	}

	// start accepting client connections
	// note that Start() returns as soon as the server is started
	err = server.Start()
	if err != nil {
		fmt.Printf("failed to start server: %v\n", err)
		os.Exit(1)
	}
	defer server.Stop()

	gh.Init()

	// Start the main ticker
	gh.Ticker()

	// never reach this point
	return
}
