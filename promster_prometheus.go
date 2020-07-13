package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery/file"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"gopkg.in/yaml.v2"
)

type PrometheusConfig config.Config

type Env struct {
	Name    string
	Version string
	Ips     []string
}

type App struct {
	Name        string    `json:"_name"`
	Description string    `json:"_desc"`
	Path        []string  `json:"_scrapePath"`
	Namespace   string    `json:"_namespace"`
	Scheme      string 
	Envs        []Env     `json:"_envs"`
	Ips					[] string `json:"_ips"`
}

func (cfg *PrometheusConfig) RegisterFlags() {
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
	log.Infof("Testando")
	log.Debugf("Config scrape interval %s", cfg.GlobalConfig.ScrapeInterval)
	flag.Var(&cfg.GlobalConfig.ScrapeInterval, "global-scrape-interval", "Prometheus scrape interval")
	flag.Var(&cfg.GlobalConfig.ScrapeTimeout, "global-scrape-timeout", "Prometheus scrape timeout")
	flag.Var(&cfg.GlobalConfig.EvaluationInterval, "global-evaluation-interval", "Prometheus global evaluation interval how frequently to evaluate rules")
	flag.StringVar(&scrapeMatch, "scrape-match", "", "Metrics regex filter applied on scraped targets. Commonly used in conjunction with /federate metrics endpoint")
	flag.StringVar(&evaluationInterval, "evaluation-interval", "30s", "Prometheus evaluation interval")
	flag.StringVar(&metrics.Scheme, "scrape-config-scheme", "http", "Scrape scheme, either http or https")

}

func (cfg *PrometheusConfig) String() string {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Sprintf("<error creating config string: %s>", err)
	}
	return string(b)
}

func (cfg *PrometheusConfig) PrintConfig(apps []App, nodeName string) {
	// d, err := yaml.Marshal(&cfg.GlobalConfig)
	// if err != nil {
	// 	log.Fatalf("error: %v", err)
	// }
	cfg.GlobalConfig.ScrapeInterval = model.Duration(15 * time.Second)
	cfg.GlobalConfig.ScrapeTimeout = model.Duration(15 * time.Second)
	cfg.GlobalConfig.EvaluationInterval = model.Duration(15 * time.Second)
	var prometheusConfig config.ScrapeConfig
	prometheusConfig.JobName = "prometheus"
	prometheusGroup := targetgroup.Group{Targets: []model.LabelSet{
		model.LabelSet{"__address__": model.LabelValue(nodeName)}}}
	prometheusConfig.ServiceDiscoveryConfig.StaticConfigs = append(prometheusConfig.ServiceDiscoveryConfig.StaticConfigs, &prometheusGroup)
	cfg.ScrapeConfigs = append(cfg.ScrapeConfigs, &prometheusConfig)

	for _, app := range apps {
		for _, path := range app.Path {
			var scrapeConfig config.ScrapeConfig
			scrapeConfig.JobName = path
			var address []model.LabelSet
			for _, ip := range app.Ips {
				lbValue := model.LabelValue(ip)
				label := model.LabelSet{"__address__": lbValue}
				address = append(address, label)
			}
			group1 := targetgroup.Group{Targets: address}
			scrapeConfig.ServiceDiscoveryConfig.StaticConfigs = append(scrapeConfig.ServiceDiscoveryConfig.StaticConfigs, &group1)
			cfg.ScrapeConfigs = append(cfg.ScrapeConfigs, &scrapeConfig)
		}
	}

	// alert config
	group2 := targetgroup.Group{Targets: []model.LabelSet{
		model.LabelSet{"__address__": "alertmanager:9093"}}}
	alConfig := config.AlertmanagerConfig{Scheme: "http"}
	alConfig.ServiceDiscoveryConfig.StaticConfigs = append(alConfig.ServiceDiscoveryConfig.StaticConfigs, &group2)
	cfg.AlertingConfig.AlertmanagerConfigs = append(cfg.AlertingConfig.AlertmanagerConfigs, &alConfig)

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
}
