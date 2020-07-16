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
	cfg := NewPromsterEtcd()
	promConfig := PromsterConfig{}
	// cfgPrometheus := NewPrometheusConfig()
	flag.StringVar(&logLevel, "log-level", "info", "debug, info, warning, error")
	flag.StringVar(&promConfig.PrometheusPathFile, "prometheus-dir-file", "/", "Dir file where prometheus.yml will be saved.")
	flag.BoolVar(&promConfig.ShardingEnabled, "scrape-shard-enable", false, "Enable sharding distribution among targets so that each Promster instance will scrape a different set of targets, enabling distribution of load among instances. Defaults to true.")
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
	
	promConfig.NodeName = cfg.NodeName
	go func() {
		for {
			log.Debugf("Prometheus nodes found: %s", promConfig.PromNodesName)
			log.Debugf("Scrape targets found: %s", promConfig.AppsTargets)
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		select {
		case promConfig.PromNodesName = <-cfg.nodesChan:
			log.Debugf("updated promNodes: %s", promConfig.PromNodesName)
		case promConfig.AppsTargets = <-cfg.appsChan:
			log.Infof("updated scapeTargets: %s", promConfig.PromNodesName)
			promConfig.PrintConfig()
			promConfig.ReloadPrometheus()
		}
		// err := updatePrometheusTargets(scrapeTargets, promNodes, cfg.scrapeShardingEnable)
		if err != nil {
			log.Warnf("Couldn't update Prometheus scrape targets. err=%s", err)
		}
	}
}
