package main

import (
	"log"
	"os"
	"sync"
	"github.com/zpatrick/go-config"
	"github.com/codegangsta/cli"

	"github.com/qnib/qframe-types"
	"github.com/qnib/qframe-handler-influxdb/lib"
	"github.com/qnib/qframe-collector-internal/lib"
	"github.com/qnib/qframe-filter-inventory/lib"
	"github.com/qnib/qframe-filter-grok/lib"
	"github.com/qnib/qframe-collector-tcp/lib"
	"github.com/qnib/qframe-collector-docker-events/lib"
	"github.com/qnib/qframe-filter-statsq/lib"
)

const (
	dockerHost = "unix:///var/run/docker.sock"
	dockerAPI = "v1.29"
)


func check_err(pname string, err error) {
	if err != nil {
		log.Printf("[EE] Failed to create %s plugin: %s", pname, err.Error())
		os.Exit(1)
	}
}

func Run(ctx *cli.Context) {
	// Create conf
	log.Printf("[II] Start Version: %s", ctx.App.Version)

	cfg := config.NewConfig([]config.Provider{})
	if _, err := os.Stat(ctx.String("config")); err == nil {
		log.Printf("[II] Use config file: %s", ctx.String("config"))
		cfg.Providers = append(cfg.Providers, config.NewYAMLFile(ctx.String("config")))
	} else {
		log.Printf("[II] No config file found")
	}
	cfg.Providers = append(cfg.Providers, config.NewCLI(ctx, false))
	qChan := qtypes.NewQChan()
	qChan.Broadcast()
	//////// Handlers
	// Start InfluxDB
	phi, err := qframe_handler_influxdb.New(qChan, cfg, "influxdb")
	check_err(phi.Name, err)
	go phi.Run()
	//////// Filters
	// GROK
	pfm, err := qframe_filter_grok.New(qChan, cfg, "grok")
	check_err(pfm.Name, err)
	go pfm.Run()
	// StatsD
	pfs, err := qframe_filter_statsq.New(qChan, cfg, "statsq")
	check_err(pfs.Name, err)
	go pfs.Run()
	// Inventory
	pfi, err := qframe_filter_inventory.New(qChan, cfg, "inventory")
	check_err(pfi.Name, err)
	go pfi.Run()
	//////// Collectors
	// Internal metrics
	pci, err := qframe_collector_internal.New(qChan, cfg, "internal")
	check_err(pci.Name, err)
	go pci.Run()
	// start docker-events
	pe, err := qframe_collector_docker_events.New(qChan, cfg, "docker-events")
	check_err(pe.Name, err)
	go pe.Run()
	// TCP
	pct, err := qframe_collector_tcp.New(qChan, cfg, "tcp")
	check_err(pct.Name, err)
	go pct.Run()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func main() {
	app := cli.NewApp()
	app.Name = "ETC event collector based on qframe, inspired by qcollect,logstash and fullerite"
	app.Usage = "qframe-metrics [options]"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "qframe.yml",
			Usage: "Config file, will overwrite flag default if present.",
		},
	}
	app.Action = Run
	app.Run(os.Args)
}
