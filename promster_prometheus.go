package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery/file"
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
		scrapeMatch        string
		evaluationInterval string
	)
	var metrics config.ScrapeConfig
	metrics.JobName = "metrics"
	metrics.MetricsPath = "/metrics"
	metrics.Scheme = "http"
	fileConfig := file.SDConfig{Files: []string{"/servers.json"}}
	metrics.ServiceDiscoveryConfig.FileSDConfigs = append(metrics.ServiceDiscoveryConfig.FileSDConfigs, &fileConfig)
	// scrape.Params.Set("teste", "localhost:9090")
	cfg.ScrapeConfigs = append(cfg.ScrapeConfigs, &metrics)

	group2 := targetgroup.Group{Targets: []model.LabelSet{
		model.LabelSet{"__address__": "alertmanager:9093"}}}
	alConfig := config.AlertmanagerConfig{Scheme: "http"}
	alConfig.ServiceDiscoveryConfig.StaticConfigs = append(alConfig.ServiceDiscoveryConfig.StaticConfigs, &group2)
	cfg.AlertingConfig.AlertmanagerConfigs = append(cfg.AlertingConfig.AlertmanagerConfigs, &alConfig)
	
	var remoteWrite config.RemoteWriteConfig

	cfg.RemoteWriteConfigs = append(cfg.RemoteWriteConfigs, &remoteWrite)
	log.Infof("Testando")
	log.Debugf("Config scrape interval %s", cfg.GlobalConfig.ScrapeInterval)
	flag.Var(&cfg.GlobalConfig.ScrapeInterval, "global-scrape-interval", "Prometheus scrape interval")
	flag.Var(&cfg.GlobalConfig.ScrapeTimeout, "global-scrape-timeout", "Prometheus scrape timeout")
	flag.Var(&cfg.GlobalConfig.EvaluationInterval, "global-evaluation-interval", "Prometheus global evaluation interval how frequently to evaluate rules")
	flag.StringVar(&scrapeMatch, "scrape-match", "", "Metrics regex filter applied on scraped targets. Commonly used in conjunction with /federate metrics endpoint")
	flag.StringVar(&evaluationInterval, "evaluation-interval", "30s", "Prometheus evaluation interval")
	flag.StringVar(&metrics.Scheme, "scrape-config-scheme", "http", "Scrape scheme, either http or https")
	flag.StringVar(&alConfig.Scheme, "alertmanager-config-scheme", "http", "Alertmanager scheme, either http or https")
	flag.Var(&remoteWrite.URL, "remotewrite-config-url", "Remote write url")
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
	var prometheusConfig config.ScrapeConfig
	prometheusConfig.JobName = "prometheus"
	group1 := targetgroup.Group{Targets: []model.LabelSet{
		model.LabelSet{"__address__": "localhost:9090"}}}
	prometheusConfig.ServiceDiscoveryConfig.StaticConfigs = append(prometheusConfig.ServiceDiscoveryConfig.StaticConfigs, &group1)
	cfg.ScrapeConfigs = append(cfg.ScrapeConfigs, &prometheusConfig)

	// alert config
	

	// rules files
	cfg.RuleFiles = []string{"/rules.yml", "/etc/prometheus/rules-l1.yml", "/etc/prometheus/rules-ln.yml", "/etc/prometheus/alert-rules.yml"}
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
