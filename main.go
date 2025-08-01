package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/kc2g-flex-tools/flexclient"
	"github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"
)

var cfg struct {
	RadioIP          string
	Station          string
	Slice            string
	Headless         bool
	SliceCreateParms string
	Listen           string
	Profile          string
	LogLevel         string
	ChkVFOMode       string
	Metering         bool
	UDPPort          int
}

func init() {
	flag.StringVar(&cfg.RadioIP, "radio", ":discover:", "radio IP address or discovery spec")
	flag.IntVar(&cfg.UDPPort, "udp-port", 0, "udp port to listen for VITA packets (0: random free port)")
	flag.StringVar(&cfg.Station, "station", "Flex", "station name to bind to or create")
	flag.StringVar(&cfg.Slice, "slice", "A", "slice letter to control")
	flag.BoolVar(&cfg.Headless, "headless", false, "run in headless mode")
	flag.StringVar(&cfg.Listen, "listen", ":4532", "hamlib listen [address]:port")
	flag.StringVar(&cfg.Profile, "profile", "", "global profile to load on startup for -headless mode")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "minimum level of messages to log to console")
	flag.StringVar(&cfg.ChkVFOMode, "chkvfo-mode", "new", "chkvfo syntax (old,new)")
	flag.BoolVar(&cfg.Metering, "metering", true, "support reading meters from radio")
}

var fc *flexclient.FlexClient
var hamlib *HamlibServer = NewHamlibServer()
var ClientID string
var ClientUUID string
var SliceIdx string

func createClient(ctx context.Context) error {
	log.Info().Msg("Registering client")
	res, err := fc.SendAndWaitContext(ctx, "client gui")
	if err != nil {
		return err
	}
	if res.Error != 0 {
		return fmt.Errorf("client gui failed: %08X", res.Error)
	}
	ClientUUID = res.Message
	ClientID = "0x" + fc.ClientID()

	if _, err := fc.SendAndWaitContext(ctx, "client program Hamlib-Flex"); err != nil {
		return err
	}
	if _, err := fc.SendAndWaitContext(ctx, "client station "+strings.ReplaceAll(cfg.Station, " ", "\x7f")); err != nil {
		return err
	}

	log.Info().Str("handle", ClientID).Msg("Got client handle")

	if cfg.Profile != "" {
		res, err := fc.SendAndWaitContext(ctx, "profile global load "+cfg.Profile)
		if err != nil {
			return err
		}
		if res.Error != 0 {
			log.Printf("Profile load failed: %08X (typo?)", res.Error)
		} else {
			log.Printf("Loaded profile %s", cfg.Profile)
		}
	}
	return nil
}

func bindClient(ctx context.Context) error {
	log.Info().Str("station", cfg.Station).Msg("Waiting for station")

	clients := make(chan flexclient.StateUpdate)
	sub := fc.Subscribe(flexclient.Subscription{Prefix: "client ", Updates: clients})
	cmdResult := fc.SendNotify("sub client all")

	var found, cmdComplete bool

	for !found || !cmdComplete {
		select {
		case <-ctx.Done():
			cmdResult.Close()
			fc.Unsubscribe(sub)
			return ctx.Err()
		case upd := <-clients:
			if upd.CurrentState["station"] == cfg.Station {
				ClientID = strings.TrimPrefix(upd.Object, "client ")
				ClientUUID = upd.CurrentState["client_id"]
				found = true
			}
		case <-cmdResult.C:
			cmdComplete = true
		}
	}
	cmdResult.Close()

	fc.Unsubscribe(sub)

	log.Info().Str("client_id", ClientID).Str("uuid", ClientUUID).Msg("Found client")

	if _, err := fc.SendAndWaitContext(ctx, "client bind client_id="+ClientUUID); err != nil {
		return err
	}
	return nil
}

func findSlice(ctx context.Context) error {
	log.Info().Str("slice_id", cfg.Slice).Msg("Looking for slice")
	slices := make(chan flexclient.StateUpdate)
	sub := fc.Subscribe(flexclient.Subscription{Prefix: "slice ", Updates: slices})
	cmdResult := fc.SendNotify("sub slice all")

	var found, cmdComplete bool

	for !found || !cmdComplete {
		select {
		case <-ctx.Done():
			cmdResult.Close()
			fc.Unsubscribe(sub)
			return ctx.Err()
		case upd := <-slices:
			if upd.CurrentState["index_letter"] == cfg.Slice && upd.CurrentState["client_handle"] == ClientID {
				SliceIdx = strings.TrimPrefix(upd.Object, "slice ")
				found = true
			}
		case <-cmdResult.C:
			cmdComplete = true
		}
	}
	cmdResult.Close()
	fc.Unsubscribe(sub)
	log.Info().Str("slice_idx", SliceIdx).Msg("Found slice")
	return nil
}

func main() {
	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{
			Out: os.Stderr,
		},
	).With().Timestamp().Logger()

	flag.Parse()

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Str("level", cfg.LogLevel).Msg("Unknown log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	if cfg.Profile != "" && !cfg.Headless {
		log.Fatal().Msg("-profile doesn't make sense without -headless")
	}

	fc, err = flexclient.NewFlexClient(cfg.RadioIP)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create FlexClient")
	}

	// Create root context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Run flexclient
	wg.Add(1)
	go func() {
		defer wg.Done()
		fc.Run()
		log.Info().Msg("FlexClient exited")
		cancel()
	}()

	// Handle signals
	go func() {
		select {
		case sig := <-sigChan:
			log.Info().Str("signal", sig.String()).Msg("Received signal, shutting down")
			cancel()
		case <-ctx.Done():
		}
	}()

	// Set up hamlib server
	err = hamlib.Listen(cfg.Listen)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start hamlib server")
	}

	// Initialize client and slice
	if cfg.Headless {
		if err := createClient(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to create client")
		}
	} else {
		if err := bindClient(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to bind client")
		}
	}

	if err := findSlice(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to find slice")
	}

	// Subscribe to updates
	if _, err := fc.SendAndWaitContext(ctx, "sub radio all"); err != nil {
		log.Error().Err(err).Msg("Failed to subscribe to radio updates")
	}
	if _, err := fc.SendAndWaitContext(ctx, "sub tx all"); err != nil {
		log.Error().Err(err).Msg("Failed to subscribe to tx updates")
	}
	if _, err := fc.SendAndWaitContext(ctx, "sub atu all"); err != nil {
		log.Error().Err(err).Msg("Failed to subscribe to atu updates")
	}

	if cfg.Metering {
		enableMetering(ctx, fc)
	}

	// Run hamlib server
	wg.Add(1)
	go func() {
		defer wg.Done()
		hamlib.Run(ctx)
		log.Info().Msg("Hamlib server exited")
	}()

	// Wait for context cancellation
	<-ctx.Done()
	log.Info().Msg("Shutting down...")

	// Close flexclient to trigger shutdown
	fc.Close()

	// Wait for all goroutines to finish
	wg.Wait()
	log.Info().Msg("Shutdown complete")
}
