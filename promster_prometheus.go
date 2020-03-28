package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"gopkg.in/yaml.v2"
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
	flag.Var(&cfg.GlobalConfig.ScrapeInterval, "global-scrape-interval", "Prometheus scrape interval")
	flag.Var(&cfg.GlobalConfig.ScrapeTimeout, "global-scrape-timeout", "Prometheus scrape timeout")
	flag.Var(&cfg.GlobalConfig.EvaluationInterval, "global-evaluation-interval", "Prometheus global evaluation interval how frequently to evaluate rules")
	flag.StringVar(&scrapeMatch, "scrape-match", "", "Metrics regex filter applied on scraped targets. Commonly used in conjunction with /federate metrics endpoint")
	flag.BoolVar(&scrapeShardingEnable, "scrape-shard-enable", false, "Enable sharding distribution among targets so that each Promster instance will scrape a different set of targets, enabling distribution of load among instances. Defaults to true.")
	flag.StringVar(&evaluationInterval, "evaluation-interval", "30s", "Prometheus evaluation interval")
	flag.StringVar(&scheme, "scheme", "http", "Scrape scheme, either http or https")

}

func (cfg *PrometheusConfig) String() string {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Sprintf("<error creating config string: %s>", err)
	}
	return string(b)
}

func (cfg *PrometheusConfig) PrintConfig() {
	// d, err := yaml.Marshal(&cfg.GlobalConfig)
	// if err != nil {
	// 	log.Fatalf("error: %v", err)
	// }
	var scrape config.ScrapeConfig
	scrape.JobName = "prometheus"
	var scrape2 config.ScrapeConfig
	scrape2.JobName = "teste"
	// scrape.Params.Set("teste", "localhost:9090")
	group1 := targetgroup.Group{Targets: []model.LabelSet{
		model.LabelSet{"__address__": "localhost:9090"}}}
	scrape.ServiceDiscoveryConfig.StaticConfigs = append(scrape.ServiceDiscoveryConfig.StaticConfigs, &group1)
	cfg.ScrapeConfigs = append(cfg.ScrapeConfigs, &scrape)
	cfg.ScrapeConfigs = append(cfg.ScrapeConfigs, &scrape2)
	d := cfg.String()
	fmt.Printf("--- m dump:\n%s\n\n", d)
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
