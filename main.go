package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/arodland/flexclient"
)

var cfg struct {
	RadioIP  string
	Station  string
	Slice    string
	Headless bool
	Listen   string
}

func init() {
	flag.StringVar(&cfg.RadioIP, "radio", "192.168.1.67", "radio IP address")
	flag.StringVar(&cfg.Station, "station", "Flex", "station name to bind to or create")
	flag.StringVar(&cfg.Slice, "slice", "A", "slice letter to control")
	flag.BoolVar(&cfg.Headless, "headless", false, "run in headless mode")
	flag.StringVar(&cfg.Listen, "listen", ":4532", "hamlib listen [address]:port")
}

var fc *flexclient.FlexClient
var hamlib *HamlibServer
var ClientID string
var ClientUUID string
var SliceIdx string

func createClient() {
	fmt.Println("Registering client")
	res := fc.SendAndWait("client gui")
	if res.Error != 0 {
		panic(res)
	}
	ClientUUID = res.Message
	ClientID = fc.ClientID()

	fc.SendAndWait("client program Hamlib-Flex")
	fc.SendAndWait("client station " + cfg.Station)
}

func bindClient() {
	fmt.Println("Waiting for station:", cfg.Station)

	clients := make(chan flexclient.StateUpdate)
	sub := fc.Subscribe(flexclient.Subscription{"client ", clients})
	cmdResult := fc.SendNotify("sub client all")

	var found, cmdComplete bool

	for !(found && cmdComplete) {
		select {
		case upd := <-clients:
			if upd.CurrentState["station"] == cfg.Station {
				ClientID = strings.TrimPrefix(upd.Object, "client ")
				ClientUUID = upd.CurrentState["client_id"]
				found = true
			}
		case <-cmdResult:
			cmdComplete = true
		}
	}

	fc.Unsubscribe(sub)

	fmt.Println("Found client ID", ClientID, "UUID", ClientUUID)

	fc.SendAndWait("client bind client_id=" + ClientUUID)
}

func findSlice() {
	fmt.Println("Looking for slice:", cfg.Slice)
	slices := make(chan flexclient.StateUpdate)
	sub := fc.Subscribe(flexclient.Subscription{"slice ", slices})
	cmdResult := fc.SendNotify("sub slice all")

	var found, cmdComplete bool

	for !(found && cmdComplete) {
		select {
		case upd := <-slices:
			if upd.CurrentState["index_letter"] == cfg.Slice && upd.CurrentState["client_handle"] == ClientID {
				SliceIdx = strings.TrimPrefix(upd.Object, "slice ")
				found = true
			}
		case <-cmdResult:
			cmdComplete = true
		}
	}

	fc.Unsubscribe(sub)
	fmt.Println("Found slice", SliceIdx)
}

func main() {
	flag.Parse()

	var err error
	fc, err = flexclient.NewFlexClient(cfg.RadioIP)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		fc.Run()
		wg.Done()
	}()

	hamlib, err = NewHamlibServer(cfg.Listen)
	if err != nil {
		panic(err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		_ = <-c
		fmt.Println("Exit on SIGINT")
		fc.Close()
		hamlib.Close()
	}()

	if cfg.Headless {
		createClient()
	} else {
		bindClient()
	}
	findSlice()

	RegisterHandlers()
	hamlib.Run()

	wg.Wait()
}
