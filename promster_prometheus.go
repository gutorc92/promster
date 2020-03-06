package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/prometheus/config"
	"gopkg.in/yaml.v1"
)

type PrometheusConfig config.Config

func NewPrometheusConfig() *PrometheusConfig {
	var config PrometheusConfig
	return &config
}

func (cfg *PrometheusConfig) RegisterFlags(f *flag.FlagSet) {
	var (
		scrapeMatch          string
		scrapeShardingEnable bool
		evaluationInterval   string
		scheme               string
	)
	log.Infof("Testando")
	log.Debugf("Config scrape interval %s", cfg.GlobalConfig.ScrapeInterval)
	flag.Var(&cfg.GlobalConfig.ScrapeInterval, "scrape-interval", "Prometheus scrape interval")
	flag.Var(&cfg.GlobalConfig.ScrapeTimeout, "scrape-timeout", "Prometheus scrape timeout")
	flag.StringVar(&scrapeMatch, "scrape-match", "", "Metrics regex filter applied on scraped targets. Commonly used in conjunction with /federate metrics endpoint")
	flag.BoolVar(&scrapeShardingEnable, "scrape-shard-enable", false, "Enable sharding distribution among targets so that each Promster instance will scrape a different set of targets, enabling distribution of load among instances. Defaults to true.")
	flag.StringVar(&evaluationInterval, "evaluation-interval", "30s", "Prometheus evaluation interval")
	flag.StringVar(&scheme, "scheme", "http", "Scrape scheme, either http or https")

}

func (cfg *PrometheusConfig) PrintConfig() {
	d, err := yaml.Marshal(&cfg.GlobalConfig)
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
	fmt.Println("Global config", cfg.GlobalConfig.ScrapeInterval)
}
