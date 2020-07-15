package main

import (
	// "encoding/json"
	"flag"
	// "fmt"
	// "io/ioutil"
	"time"

	// "github.com/serialx/hashring"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {
	var logLevel string
	var sharding bool
	cfg := NewPromsterEtcd()
	// cfgPrometheus := NewPrometheusConfig()
	flag.StringVar(&logLevel, "log-level", "info", "debug, info, warning, error")
	flag.BoolVar(&sharding, "scrape-shard-enable", false, "Enable sharding distribution among targets so that each Promster instance will scrape a different set of targets, enabling distribution of load among instances. Defaults to true.")
	cfg.RegisterFlags()
	// cfgPrometheus.RegisterFlags(flagSetPromster)
	flag.Parse()
	ll, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Errorf("Not able to parse log level string. Setting default level: info.")
		ll = logrus.InfoLevel
	}
	logrus.SetLevel(ll)
	// cfgPrometheus.PrintConfig()
	cfg.CheckFlags()
	log.Infof("====Starting Promster====")
	cfg.InitPromsterEtcd()

	InitWatch(cfg)

	promNodes := make([]string, 0)
	appsTargets := make([]App, 0)
	go func() {
		for {
			log.Debugf("Prometheus nodes found: %s", promNodes)
			log.Debugf("Scrape targets found: %s", appsTargets)
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		select {
		case promNodes = <-cfg.nodesChan:
			log.Debugf("updated promNodes: %s", promNodes)
		case appsTargets = <-cfg.appsChan:
			log.Infof("updated scapeTargets: %s", appsTargets)
			cfgPrometheus := PrometheusConfig{}
			cfgPrometheus.PrintConfig(appsTargets, cfg.nodeName, sharding)
			cfgPrometheus.ReloadPrometheus()

		}
		// err := updatePrometheusTargets(scrapeTargets, promNodes, cfg.scrapeShardingEnable)
		if err != nil {
			log.Warnf("Couldn't update Prometheus scrape targets. err=%s", err)
		}
	}
}
