package main

import (
	"encoding/json"
	"context"
	"flag"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	etcdregistry "github.com/flaviostutz/etcd-registry/etcd-registry"
)

type PromsterEtcd struct {
	URLRegistry       string
	Base              string
	ServiceName       string
	ServiceTTL        int
	URLScrape         string
	scrapeEtcdPath    string
	SettingsPath      chan string
	cliScrape         *clientv3.Client
	cliRegistry				*clientv3.Client
	appsChan          chan []App
	settingsChan      chan []Setting
	nodesChan         chan []string
	Sharding          bool
	NodeName          string
}

func (cfg *PromsterEtcd) RegisterFlags() {
	flag.StringVar(&cfg.URLRegistry, "registry-etcd-url", "", "ETCD URLs. ex: http://etcd0:2379")
	flag.StringVar(&cfg.Base, "registry-etcd-base", "/registry", "ETCD base path for services")
	flag.StringVar(&cfg.ServiceName, "registry-service-name", "", "Prometheus cluster service name. Ex.: proml1")
	flag.IntVar(&cfg.ServiceTTL, "registry-node-ttl", -1, "Node registration TTL in ETCD. After killing Promster instance, it will vanish from ETCD registry after this time")
	flag.StringVar(&cfg.URLScrape, "scrape-etcd-url", "", "ETCD URLs for scrape source server. If empty, will be the same as --etcd-url. ex: http://etcd0:2379")
	flag.StringVar(&cfg.scrapeEtcdPath, "scrape-etcd-path", "", "Base ETCD path for getting servers to be scrapped")
}

func (cfg *PromsterEtcd) hasEtcdRegistry() bool {
	if cfg.URLRegistry != "" {
		return true
	}
	return false
}

func (cfg *PromsterEtcd) CheckFlags() {
	log.Infof("==== Parssing Etcd Variables ====")
	if cfg.URLScrape == "" {
		panic("--scrape-etcd-url should be defined")
	}
	log.Debugf("Etcd URL Scrape: %s", cfg.URLScrape)

	if cfg.URLRegistry != "" {
		if cfg.Base == "" {
			panic("--etcd-base should be defined")
		}
		log.Debugf("Etcd Base: %s", cfg.Base)
		if cfg.ServiceName == "" {
			panic("--etcd-service-name should be defined")
		}
		log.Debugf("Etcd Service Name: %s", cfg.ServiceName)
		if cfg.ServiceTTL == -1 {
			panic("--etcd-node-ttl should be defined")
		}
		log.Debugf("Etcd Service TTL: %d", cfg.ServiceTTL)
	}
	if cfg.scrapeEtcdPath == "" {
		log.Debugf("Scrape Etcd Path: %s", cfg.scrapeEtcdPath)
	}
	log.Debugf("Scrape Etcd Path: %s", cfg.scrapeEtcdPath)
}

func NewPromsterEtcd() *PromsterEtcd {
	var promster PromsterEtcd
	return &promster
}

func (cfg *PromsterEtcd) servicePath() string {
	servicePath := fmt.Sprintf("%s/%s/", cfg.Base, cfg.ServiceName)
	return servicePath
}

func (cfg *PromsterEtcd) InitClients () {
	endpointsScrape := strings.Split(cfg.URLScrape, ",")
	cliScrape, err := clientv3.New(clientv3.Config{Endpoints: endpointsScrape, DialTimeout: 10 * time.Second})
	if err != nil {
		log.Errorf("Could not initialize ETCD client. err=%s", err)
		panic(err)
	}
	cfg.cliScrape = cliScrape
	endpointsRegistry := strings.Split(cfg.URLRegistry, ",")
	cliRegistry, err := clientv3.New(clientv3.Config{Endpoints: endpointsRegistry, DialTimeout: 10 * time.Second})
	if err != nil {
		log.Errorf("Could not initialize ETCD client. err=%s", err)
		panic(err)
	}
	cfg.cliRegistry = cliRegistry
}

func (cfg *PromsterEtcd) InitPromsterEtcd() {
	cfg.InitClients()
	cfg.nodesChan = make(chan []string, 0)
	cfg.appsChan = make(chan []App, 0)
	cfg.NodeName = getSelfNodeName()
}

func InitWatch(cfg *PromsterEtcd) {
	log.Infof("Starting to watch scrape targets...")
	go watchTargets(cfg)
	log.Infof("Keeping self node registered on ETCD...")
	go registerNode(cfg)
	log.Infof("Starting to watch registered prometheus nodes...")
	go watchNodes(cfg)
}

func registerNode(cfg *PromsterEtcd) {
	endpointsRegistry := strings.Split(cfg.URLRegistry, ",")
	registry, err := etcdregistry.NewEtcdRegistry(endpointsRegistry, cfg.Base, 10*time.Second)
	if err != nil {
		panic(err)
	}
	node := etcdregistry.Node{}
	node.Name = cfg.NodeName
	log.Debugf("Registering Prometheus instance on ETCD registry. service=%s; node=%s", cfg.ServiceName, node)
	err = registry.RegisterNode(context.TODO(), cfg.ServiceName, node, time.Duration(cfg.ServiceTTL)*time.Second)
	if err != nil {
		panic(err)
	}
}

func watchNodes(cfg *PromsterEtcd) {
	watchChan := cfg.cliRegistry.Watch(context.TODO(), cfg.servicePath(), clientv3.WithPrefix())
	for {
		log.Debugf("Registered nodes updated")
		rsp, err0 := cfg.cliRegistry.Get(context.TODO(), cfg.servicePath(), clientv3.WithPrefix())
		if err0 != nil {
			log.Warnf("Error retrieving service nodes. err=%s", err0)
		}

		if len(rsp.Kvs) == 0 {
			log.Debugf("No services nodes were found under %s", cfg.servicePath())

		} else {
			promNodes := make([]string, 0)
			for _, kv := range rsp.Kvs {
				promNodes = append(promNodes, path.Base(string(kv.Key)))
			}
			cfg.nodesChan <- promNodes
			log.Debugf("Found registered nodes %s", promNodes)
		}
		<-watchChan
	}
}

func getSelfNodeName() string {
	hostip, err := ExecShell("ip route get 8.8.8.8 | grep -oE 'src ([0-9\\.]+)' | cut -d ' ' -f 2")
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s:9090", strings.TrimSpace(hostip))
}

func loadAppSettings(cfg *PromsterEtcd, app *App) {
	settingsPath := fmt.Sprintf("/settings/%s", app.Namespace)
	rsp, err0 := cfg.cliScrape.Get(context.TODO(), settingsPath, clientv3.WithPrefix())
	if err0 != nil {
		log.Warnf("Error retrieving setting scrape targets. err=%s", err0)
	}
	log.Infof("Settings found: %d", len(rsp.Kvs))
	// cfg.SettingsPath <- settingsPath
	for _, kv := range rsp.Kvs {
		log.Infof("Settings found %s", string(kv.Value))
		var setting Setting
		err := json.Unmarshal([]byte(kv.Value), &setting)
		if err != nil {
			log.Infof("Error unmarshal settings %s", err)
		}
		log.Infof("Settins unmarshall %+v", setting)
		app.Setting = &setting
		// return &setting
	}
}

func watchTargets(cfg *PromsterEtcd) {
	log.Infof("Testando o info na funcao")
	log.Infof("Getting source scrape targets from %s", cfg.scrapeEtcdPath)
	watchChan := cfg.cliScrape.Watch(context.TODO(), cfg.scrapeEtcdPath, clientv3.WithPrefix())
	for {
		log.Infof("Source scrape targets updated")
		rsp, err0 := cfg.cliScrape.Get(context.TODO(), cfg.scrapeEtcdPath, clientv3.WithPrefix())
		if err0 != nil {
			log.Warnf("Error retrieving source scrape targets. err=%s", err0)
		}
		if len(rsp.Kvs) == 0 {
			log.Infof("No source scrape targets were found under %s", cfg.scrapeEtcdPath)

		} else {
			appsTargets := make([]App, 0)
			for _, kv := range rsp.Kvs {
				record := string(kv.Key)
				app := App{}
				app.Name = record
				log.Infof("App found %s", string(kv.Value))
				err := json.Unmarshal([]byte(kv.Value), &app)
				if err != nil {
					log.Infof("Error unmarshal app %s", err)
				} else {
					loadAppSettings(cfg, &app)
					appsTargets = append(appsTargets, app)
				}
			}
			cfg.appsChan <- appsTargets
			log.Infof("Found source scrape targets: %s", appsTargets)
		}
		<-watchChan
	}

	// log.Infof("Updating scrape targets for this shard to %s")
}
