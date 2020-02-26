package main

import (
	"fmt"
	"flag"
	"gopkg.in/yaml.v1"
	"os"
	"github.com/prometheus/prometheus/config"
)

var configPromster config.Config

func RegisterFlags(f *flag.FlagSet) {
	var (
		scrapeInterval string
		scrapeTimeout string
		scrapeMatch string
		scrapeShardingEnable bool
		evaluationInterval string
		scheme string
	)
	flag.IntVar(&configPromster.GlobalConfig.ScrapeInterval, "scrape-interval", 30, "Prometheus scrape interval")
	flag.IntVar(&configPromster.GlobalConfig.ScrapeTimeout, "scrape-timeout", 30, "Prometheus scrape timeout")
	flag.StringVar(&scrapeMatch, "scrape-match", "", "Metrics regex filter applied on scraped targets. Commonly used in conjunction with /federate metrics endpoint")
	flag.BoolVar(&scrapeShardingEnable, "scrape-shard-enable", false, "Enable sharding distribution among targets so that each Promster instance will scrape a different set of targets, enabling distribution of load among instances. Defaults to true.")
	flag.StringVar(&evaluationInterval, "evaluation-interval", "30s", "Prometheus evaluation interval")
	flag.StringVar(&scheme, "scheme", "http", "Scrape scheme, either http or https")
	flag.Parse()
	
}

func createConfig () {
	d, err := yaml.Marshal(&config.GlobalConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m dump:\n%s\n\n", string(d))
	f, err := os.Create("prometheus_test.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	l, err := f.WriteString(string(d))
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(l, "bytes written successfully")
	fmt.Println("Global config", config.GlobalConfig.ScrapeInterval)
}
