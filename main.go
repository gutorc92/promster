package main

import (
	"flag"
	"os"
	"github.com/sirupsen/logrus"
	"time"
)

var log = logrus.New()

func setLogLevel(logLevel *string) {
	switch *logLevel {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		log.SetLevel(logrus.WarnLevel)
		break
	case "error":
		log.SetLevel(logrus.ErrorLevel)
		break
	default:
		log.Infof("Setando o valor default")
		log.SetLevel(logrus.InfoLevel)
	}
}
func main() {
	createConfig()
	var cfg *PromsterEtcd
	cfg = NewPromsterEtcd()
	log.Out = os.Stdout
	logLevel := flag.String("loglevel", "info", "debug, info, warning, error")
	flagSet := flag.NewFlagSet("etcd", flag.ContinueOnError)
	cfg.RegisterFlags(flagSet)
	flag.Parse()
	setLogLevel(logLevel)
	cfg.CheckFlags()
	log.Debugf("Testandoo valor: %s ", cfg.etcdURLRegistry)
	log.Infof("====Starting Promster====")
	CreateEtcd(cfg)
	log.Debugf("Initializing ETCD client for source scrape targets")
	log.Infof("Starting to watch source scrape targets. etcdURLScrape=%s", cfg.etcdURLScrape)
	
	cfg.createTargets()

	promNodes := make([]string, 0)
	scrapeTargets := make([]SourceTarget, 0)
	go func() {
		for {
			log.Debugf("Prometheus nodes found: %s", promNodes)
			log.Debugf("Scrape targets found: %s", scrapeTargets)
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		select {
		case promNodes = <- cfg.nodesChan:
			log.Debugf("updated promNodes: %s", promNodes)
		case scrapeTargets = <- cfg.sourceTargetsChan:
			log.Debugf("updated scapeTargets: %s", scrapeTargets)
		}
		// err := updatePrometheusTargets(scrapeTargets, promNodes, scrapeShardingEnable)
		// if err != nil {
		// 	log.Warnf("Couldn't update Prometheus scrape targets. err=%s", err)
		// }
	}
}