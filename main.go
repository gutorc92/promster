package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/serialx/hashring"
	"github.com/sirupsen/logrus"
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

func updatePrometheusTargets(scrapeTargets []SourceTarget, promNodes []string, shardingEnabled bool) error {
	//Apply consistent hashing to determine which scrape endpoints will
	//be handled by this Prometheus instance
	logrus.Debugf("updatePrometheusTargets. scrapeTargets=%s, promNodes=%s", scrapeTargets, promNodes)

	ring := hashring.New(hashList(promNodes))
	selfNodeName := getSelfNodeName()
	selfScrapeTargets := make([]SourceTarget, 0)
	for _, starget := range scrapeTargets {
		hashedPromNode, ok := ring.GetNode(stringSha512(starget.Targets[0]))
		if !ok {
			return fmt.Errorf("Couldn't get prometheus node for %s in consistent hash", starget.Targets[0])
		}
		logrus.Debugf("Target %s - Prometheus %x", starget, hashedPromNode)
		hashedSelf := stringSha512(selfNodeName)
		if !shardingEnabled || hashedSelf == hashedPromNode {
			logrus.Debugf("Target %s - Prometheus %s", starget, selfNodeName)
			selfScrapeTargets = append(selfScrapeTargets, starget)
		}
	}

	//generate json file
	contents, err := json.Marshal(selfScrapeTargets)
	if err != nil {
		return err
	}
	logrus.Debugf("Writing /servers.json: '%s'", string(contents))
	err = ioutil.WriteFile("/servers.json", contents, 0666)
	if err != nil {
		return err
	}

	//force Prometheus to update its configuration live
	_, err = ExecShell("wget --post-data='' http://localhost:9090/-/reload -O -")
	if err != nil {
		return err
	}
	output, err0 := ExecShell("kill -HUP $(ps | grep prometheus | awk '{print $1}' | head -1)")
	if err0 != nil {
		logrus.Warnf("Could not reload Prometheus configuration. err=%s. output=%s", err0, output)
	}

	return nil
}

func main() {
	var cfg *PromsterEtcd
	var cfgPrometheus *PrometheusConfig
	cfg = NewPromsterEtcd()
	cfgPrometheus = NewPrometheusConfig()
	log.Out = os.Stdout
	logLevel := flag.String("loglevel", "info", "debug, info, warning, error")
	flagSet := flag.NewFlagSet("etcd", flag.ContinueOnError)
	flagSetPromster := flag.NewFlagSet("prometheus", flag.ContinueOnError)
	cfg.RegisterFlags(flagSet)
	cfgPrometheus.RegisterFlags(flagSetPromster)
	flag.Parse()
	cfgPrometheus.PrintConfig()
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
		case promNodes = <-cfg.nodesChan:
			log.Debugf("updated promNodes: %s", promNodes)
		case scrapeTargets = <-cfg.sourceTargetsChan:
			log.Debugf("updated scapeTargets: %s", scrapeTargets)
		}
		err := updatePrometheusTargets(scrapeTargets, promNodes, cfg.scrapeShardingEnable)
		if err != nil {
			log.Warnf("Couldn't update Prometheus scrape targets. err=%s", err)
		}
	}
}
