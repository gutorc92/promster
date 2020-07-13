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

// func updatePrometheusTargets(scrapeTargets []SourceTarget, promNodes []string, shardingEnabled bool) error {
// 	//Apply consistent hashing to determine which scrape endpoints will
// 	//be handled by this Prometheus instance
// 	logrus.Debugf("updatePrometheusTargets. scrapeTargets=%s, promNodes=%s", scrapeTargets, promNodes)

// 	ring := hashring.New(hashList(promNodes))
// 	selfNodeName := getSelfNodeName()
// 	selfScrapeTargets := make([]SourceTarget, 0)
// 	for _, starget := range scrapeTargets {
// 		hashedPromNode, ok := ring.GetNode(stringSha512(starget.Targets[0]))
// 		if !ok {
// 			return fmt.Errorf("Couldn't get prometheus node for %s in consistent hash", starget.Targets[0])
// 		}
// 		logrus.Debugf("Target %s - Prometheus %x", starget, hashedPromNode)
// 		hashedSelf := stringSha512(selfNodeName)
// 		if !shardingEnabled || hashedSelf == hashedPromNode {
// 			logrus.Debugf("Target %s - Prometheus %s", starget, selfNodeName)
// 			selfScrapeTargets = append(selfScrapeTargets, starget)
// 		}
// 	}

// 	//generate json file
// 	contents, err := json.Marshal(selfScrapeTargets)
// 	if err != nil {
// 		return err
// 	}
// 	logrus.Debugf("Writing /servers.json: '%s'", string(contents))
// 	err = ioutil.WriteFile("/servers.json", contents, 0666)
// 	if err != nil {
// 		return err
// 	}

// 	//force Prometheus to update its configuration live
// 	_, err = ExecShell("wget --post-data='' http://localhost:9090/-/reload -O -")
// 	if err != nil {
// 		return err
// 	}
// 	output, err0 := ExecShell("kill -HUP $(ps | grep prometheus | awk '{print $1}' | head -1)")
// 	if err0 != nil {
// 		logrus.Warnf("Could not reload Prometheus configuration. err=%s. output=%s", err0, output)
// 	}

// 	return nil
// }

func main() {
	var logLevel string
	cfg := NewPromsterEtcd()
	// cfgPrometheus := NewPrometheusConfig()
	flag.StringVar(&logLevel, "log-level", "info", "debug, info, warning, error")
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

		}
		// err := updatePrometheusTargets(scrapeTargets, promNodes, cfg.scrapeShardingEnable)
		if err != nil {
			log.Warnf("Couldn't update Prometheus scrape targets. err=%s", err)
		}
	}
}
