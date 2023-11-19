package main

import (
	"flag"

	"github.com/faustuzas/tcha/src/cliclient"
)

var gatewayAddress = flag.String("a", "localhost:9000", "gateway address to connect to")

func main() {
	cliclient.Run(cliclient.Params{
		Configuration: cliclient.Configuration{
			GatewayAddress: *gatewayAddress,
		},
	})
}
