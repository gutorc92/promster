package main

import (
	"context"
	"flag"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	etcdregistry "github.com/flaviostutz/etcd-registry/etcd-registry"
)

type SourceTarget struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels,omitempty"`
}

type PromsterEtcd struct {
	URLRegistry       string
	Base              string
	ServiceName       string
	ServiceTTL        int
	URLScrape         string
	scrapeEtcdPath    string
	cliScrape         *clientv3.Client
	sourceTargetsChan chan []SourceTarget
	nodesChan         chan []string
	Sharding          bool
	nodeName          string
}

func (cfg *PromsterEtcd) RegisterFlags(f *flag.FlagSet) {
	flag.StringVar(&cfg.URLRegistry, "registry-etcd-url", "", "ETCD URLs. ex: http://etcd0:2379")
	flag.StringVar(&cfg.Base, "registry-etcd-base", "/registry", "ETCD base path for services")
	flag.StringVar(&cfg.ServiceName, "registry-service-name", "", "Prometheus cluster service name. Ex.: proml1")
	flag.BoolVar(&cfg.Sharding, "scrape-shard-enable", false, "Enable sharding distribution among targets so that each Promster instance will scrape a different set of targets, enabling distribution of load among instances. Defaults to true.")
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
		panic("--etcd-url-scrape should be defined")
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
		panic("--scrape-etcd-path should be defined")
	}
	log.Debugf("Scrape Etcd Path: %s", cfg.scrapeEtcdPath)
}

func NewPromsterEtcd() *PromsterEtcd {
	var promster PromsterEtcd
	return &promster
}

func (cfg *PromsterEtcd) createRegistry() *clientv3.Client {
	endpointsRegistry := strings.Split(cfg.URLRegistry, ",")
	cliRegistry, err := clientv3.New(clientv3.Config{Endpoints: endpointsRegistry, DialTimeout: 10 * time.Second})
	if err != nil {
		log.Errorf("Could not initialize ETCD client. err=%s", err)
		panic(err)
	}
	return cliRegistry
}

func (cfg *PromsterEtcd) servicePath() string {
	servicePath := fmt.Sprintf("%s/%s/", cfg.Base, cfg.ServiceName)
	return servicePath
}

func (cfg *PromsterEtcd) InitWatch() {
	cfg.nodesChan = make(chan []string, 0)
	cfg.nodeName = getSelfNodeName()
	if cfg.hasEtcdRegistry() {
		log.Infof("Keeping self node registered on ETCD...")
		go cfg.keepSelfNodeRegistered()
		log.Infof("Starting to watch registered prometheus nodes...")
		// go watchRegisteredNodes(cliRegistry, servicePath, cfg.nodesChan)
		go cfg.watchRegisteredNodes()
	} else {
		go func() {
			cfg.nodesChan <- []string{cfg.nodeName}
		}()
	}
}

func (cfg *PromsterEtcd) keepSelfNodeRegistered() {
	endpointsRegistry := strings.Split(cfg.URLRegistry, ",")
	registry, err := etcdregistry.NewEtcdRegistry(endpointsRegistry, cfg.Base, 10*time.Second)
	if err != nil {
		panic(err)
	}
	node := etcdregistry.Node{}
	node.Name = cfg.nodeName
	log.Debugf("Registering Prometheus instance on ETCD registry. service=%s; node=%s", cfg.ServiceName, node)
	err = registry.RegisterNode(context.TODO(), cfg.ServiceName, node, time.Duration(cfg.ServiceTTL)*time.Second)
	if err != nil {
		panic(err)
	}
}

func (cfg *PromsterEtcd) createTargets() {
	log.Debugf("Initializing ETCD client for source scrape targets")
	log.Infof("Starting to watch source scrape targets. etcdURLScrape=%s", cfg.URLScrape)
	endpointsScrape := strings.Split(cfg.URLScrape, ",")
	cliScrape, err := clientv3.New(clientv3.Config{Endpoints: endpointsScrape, DialTimeout: 10 * time.Second})
	if err != nil {
		log.Errorf("Could not initialize ETCD client. err=%s", err)
		panic(err)
	}
	log.Infof("Etcd client initialized for scrape")
	cfg.sourceTargetsChan = make(chan []SourceTarget, 0)
	go watchSourceScrapeTargets(cliScrape, cfg.scrapeEtcdPath, cfg.sourceTargetsChan)
}

func (cfg *PromsterEtcd) watchRegisteredNodes() {
	cli := cfg.createRegistry()
	watchChan := cli.Watch(context.TODO(), cfg.servicePath(), clientv3.WithPrefix())
	for {
		log.Debugf("Registered nodes updated")
		rsp, err0 := cli.Get(context.TODO(), cfg.servicePath(), clientv3.WithPrefix())
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

func watchSourceScrapeTargets(cli *clientv3.Client, sourceTargetsPath string, sourceTargetsChan chan []SourceTarget) {
	log.Debugf("Getting source scrape targets from %s", sourceTargetsPath)

	watchChan := cli.Watch(context.TODO(), sourceTargetsPath, clientv3.WithPrefix())
	for {
		log.Debugf("Source scrape targets updated")
		rsp, err0 := cli.Get(context.TODO(), sourceTargetsPath, clientv3.WithPrefix())
		if err0 != nil {
			log.Warnf("Error retrieving source scrape targets. err=%s", err0)
		}

		if len(rsp.Kvs) == 0 {
			log.Debugf("No source scrape targets were found under %s", sourceTargetsPath)

		} else {
			sourceTargets := make([]SourceTarget, 0)
			for _, kv := range rsp.Kvs {
				record := string(kv.Key)
				targetAddress := path.Base(record)
				serviceName := path.Base(path.Dir(record))
				sourceTargets = append(sourceTargets, SourceTarget{Labels: map[string]string{"prsn": serviceName}, Targets: []string{targetAddress}})
			}
			sourceTargetsChan <- sourceTargets
			log.Debugf("Found source scrape targets: %s", sourceTargets)
		}
		<-watchChan
	}

	// log.Infof("Updating scrape targets for this shard to %s")
}
