package main

import (
	"flag"
)

type PromsterEtcd struct {
	etcdURLRegistry string
	etcdBase        *string
	etcdServiceName string
	etcdServiceTTL  int
	etcdURLScrape   string
	scrapeEtcdPath  string
}

func (cfg *PromsterEtcd) RegisterFlags(f *flag.FlagSet) {
	flag.StringVar(&cfg.etcdURLRegistry, "registry-etcd-url", "", "ETCD URLs. ex: http://etcd0:2379")
	// cfg.etcdBase = flag.String("registry-etcd-base", "/registry", "ETCD base path for services")
	// cfg.etcdServiceName = flag.String("registry-service-name", "", "Prometheus cluster service name. Ex.: proml1")
	// cfg.etcdServiceTTL = flag.Int("registry-node-ttl", -1, "Node registration TTL in ETCD. After killing Promster instance, it will vanish from ETCD registry after this time")
	// cfg.etcdURLScrape = flag.String("scrape-etcd-url", "", "ETCD URLs for scrape source server. If empty, will be the same as --etcd-url. ex: http://etcd0:2379")
	// cfg.scrapeEtcdPath = flag.String("scrape-etcd-path", "", "Base ETCD path for getting servers to be scrapped")
}

func NewPromsterEtcd() *PromsterEtcd {
	var promster PromsterEtcd
	return &promster
}
