package main

import (
	"flag"

	"github.com/faustuzas/occa/src/cliclient"
)

var gatewayAddress = flag.String("a", "localhost:9000", "gateway address to connect to")

func main() {
	flag.Parse()

	cliclient.Run(cliclient.Params{
		Configuration: cliclient.Configuration{
			GatewayAddress: *gatewayAddress,
		},
	})
}
