package main

import (
	"flag"
	"testing"
)

// var pkgdir = flag.String("pkgdir", "", "dir of package containing embedded files")
flagSet := flag.NewFlagSet("teste", ContinueOnError)

func TestCreateEtcConfig(t *testing.T) {
	var cfg *PromsterEtcd
	cfg = NewPromsterEtcd()
	cfg.RegisterFlags(&flagSet)
	t.Error()
}
