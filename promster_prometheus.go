package main

import (
	"fmt"
	"os"
	"time"
	"net/url"
	"github.com/serialx/hashring"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	config_util "github.com/prometheus/common/config"
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
	Setting			*Setting
}

type Setting struct {
	EtcdService  string `json:"_etcdService"`
	Namespace    string `json:"_namespace"`
	RemoteWrite  string `json:"_remoteWrite"`
	Alertmanager string `json:"_alertManager"`
}

type PromsterConfig struct {
	AppsTargets        []App
	SettingsTargets    []Setting
	PromNodesName      []string
	NodeName           string
	ShardingEnabled    bool
	PrometheusPathFile string
}

func (cfg *PrometheusConfig) String() string {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Sprintf("<error creating config string: %s>", err)
	}
	return string(b)
}

func (cfg *PromsterConfig) ReloadPrometheus() error {
	//force Prometheus to update its configuration live
	_, err := ExecShell("wget --post-data='' http://localhost:9090/-/reload -O -")
	if err != nil {
		return err
	}
	output, err0 := ExecShell("kill -HUP $(ps | grep prometheus | awk '{print $1}' | head -1)")
	if err0 != nil {
		log.Warnf("Could not reload Prometheus configuration. err=%s. output=%s", err0, output)
	}

	return nil
}

func (cf *PromsterConfig) IPsToLabelSet(ips []string) []model.LabelSet {
	var address []model.LabelSet
	ring := hashring.New(hashList(cf.PromNodesName))
	for _, ip := range ips {
		lbValue := model.LabelValue(ip)
		label := model.LabelSet{"__address__": lbValue}
		hashedPromNode, ok := ring.GetNode(stringSha512(ip))
		if !ok {
			log.Errorf("Couldn't get prometheus node for %s in consistent hash", ip)
		}
		log.Debugf("Target %s - Prometheus %x", ip, cf.NodeName)
		hashedSelf := stringSha512(cf.NodeName)
		if !cf.ShardingEnabled || hashedSelf == hashedPromNode {
			log.Debugf("Target %s - Prometheus %s", ip, cf.NodeName)
			address = append(address, label)
		}
	}
	return address
}

func (cf *PromsterConfig) PrintConfig() {
	// d, err := yaml.Marshal(&cfg.GlobalConfig)
	// if err != nil {
	// 	log.Fatalf("error: %v", err)
	// }
	cfg := PrometheusConfig{}
	cfg.GlobalConfig.ScrapeInterval = model.Duration(15 * time.Second)
	cfg.GlobalConfig.ScrapeTimeout = model.Duration(15 * time.Second)
	cfg.GlobalConfig.EvaluationInterval = model.Duration(15 * time.Second)
	var prometheusConfig config.ScrapeConfig
	prometheusConfig.JobName = "prometheus"
	prometheusGroup := targetgroup.Group{Targets: []model.LabelSet{
		model.LabelSet{"__address__": model.LabelValue(cf.NodeName)}}}
	prometheusConfig.ServiceDiscoveryConfig.StaticConfigs = append(prometheusConfig.ServiceDiscoveryConfig.StaticConfigs, &prometheusGroup)
	cfg.ScrapeConfigs = append(cfg.ScrapeConfigs, &prometheusConfig)

	for _, app := range cf.AppsTargets {
		paths := make([]string, len(app.Path))
		copy(paths, app.Path)
		if len(app.Path) == 0 {
			paths = append(paths, app.Name)
		}
		if app.Setting != nil && app.Setting.RemoteWrite != "" {
			var urlParsed config_util.URL 
			urlParsed2, err := url.Parse(app.Setting.RemoteWrite)
			if err != nil {
				log.Infof("Error on coverting url of settings: %s +v", app.Namespace, urlParsed2)	
			} else {
				log.Infof("Url parsed %+v", urlParsed)
				urlParsed.URL = urlParsed2
				remoteConfig := config.RemoteWriteConfig{
					URL: &urlParsed,
					RemoteTimeout: model.Duration(15 * time.Second),
				}
				cfg.RemoteWriteConfigs = append(cfg.RemoteWriteConfigs, &remoteConfig)
			}
		}
		for _, path := range paths {
			var scrapeConfig config.ScrapeConfig
			scrapeConfig.JobName = path
			group1 := targetgroup.Group{Targets: cf.IPsToLabelSet(app.Ips)}
			scrapeConfig.ServiceDiscoveryConfig.StaticConfigs = append(scrapeConfig.ServiceDiscoveryConfig.StaticConfigs, &group1)
			cfg.ScrapeConfigs = append(cfg.ScrapeConfigs, &scrapeConfig)
		}
	}

	// alert config
	group2 := targetgroup.Group{Targets: []model.LabelSet{
		model.LabelSet{"__address__": "alertmanager:9093"}}}
	alConfig := config.AlertmanagerConfig{Scheme: "http", APIVersion: "v2"}
	alConfig.ServiceDiscoveryConfig.StaticConfigs = append(alConfig.ServiceDiscoveryConfig.StaticConfigs, &group2)
	cfg.AlertingConfig.AlertmanagerConfigs = append(cfg.AlertingConfig.AlertmanagerConfigs, &alConfig)

	// rules files
	cfg.RuleFiles = []string{"/rules.yml", "/etc/prometheus/rules-l1.yml", "/etc/prometheus/rules-ln.yml", "/etc/prometheus/alert-rules.yml"}
	d := cfg.String()
	fmt.Printf("--- m dump:\n%s\n\n", d)
	pathPrometheus := fmt.Sprintf("%s/%s", cf.PrometheusPathFile, "prometheus.yml")
	f, err := os.Create(pathPrometheus)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	l, err := f.WriteString(string(d))
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(l, "bytes written successfully")
}
