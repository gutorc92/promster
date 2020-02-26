package main

import (
	"flag"
	"testing"
)

// var pkgdir = flag.String("pkgdir", "", "dir of package containing embedded files")

func TestCreateEtcConfig(t *testing.T) {
	flagSet := flag.NewFlagSet("teste", flag.ContinueOnError)
	var cfg *PromsterEtcd
	cfg = NewPromsterEtcd()
	cfg.RegisterFlags(flagSet)
	flag.Parse()
	t.Logf("valor de teste %s", cfg.etcdURLRegistry)
	t.Error()
}
