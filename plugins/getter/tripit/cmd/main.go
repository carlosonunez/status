package main

import (
	"os"

	tripit "github.com/carlosonunez/status/plugins/getter/tripit"
	"github.com/carlosonunez/status/pkg/pluginsdk"
)

func main() {
	os.Exit(pluginsdk.ServeGetter(tripit.New(), os.Args[1:], os.Stdin, os.Stdout))
}
