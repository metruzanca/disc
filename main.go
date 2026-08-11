package main

import "github.com/metruzanca/disc/cmd"

// version is injected at build time by GoReleaser (see .goreleaser.yaml).
var version = "dev"

func main() {
	cmd.Execute()
}
