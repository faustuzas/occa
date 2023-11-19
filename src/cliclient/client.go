package cliclient

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"

	gatewayclient "github.com/faustuzas/occa/src/gateway/client"
)

type Params struct {
	Configuration
}

type Configuration struct {
	GatewayAddress string
}

func Run(params Params) {
	app := cliClientApplication{
		Params:  params,
		gateway: gatewayclient.New(params.GatewayAddress),
		console: Console{},
	}

	steps := []func() bool{
		app.printIntro,
		app.checkGatewayConnectivity,
		app.authentication,
	}

	for _, stage := range steps {
		if !stage() {
			return
		}
	}
}

type cliClientApplication struct {
	Params

	gateway *gatewayclient.Client
	console Console
}

func (a cliClientApplication) printIntro() bool {
	a.console.PrintHeader("Welcome to OCCA!")
	return true
}

func (a cliClientApplication) checkGatewayConnectivity() bool {
	a.console.PrintRegular("Checking connection to gateway@%v ...", a.GatewayAddress)
	if err := a.gateway.CheckAvailability(a.ctx()); err != nil {
		a.console.PrintRegular("Connecting to gateway failed. Reason: %v", err)
		return false
	}
	a.console.PrintRegular("Connection to gateway successful!")
	return true
}

func (a cliClientApplication) authentication() bool {
	a.console.PrintNewLine()

	userName := a.console.Prompt("Enter your user name")
	password := a.console.PromptPassword("Enter your password")

	a.console.PrintNewLine()
	if err := a.gateway.Authenticate(a.ctx(), userName, password); err != nil {
		a.console.PrintRegular("Failed to authenticate. Reason: %v", err)
		return false
	}
	a.console.PrintRegular("Authenticated successfully. Welcome back, %v!", userName)

	return true
}

func (a cliClientApplication) ctx() context.Context {
	return context.Background()
}

type Console struct {
}

func (c Console) PrintHeader(header string) {
	fmt.Printf("\t%v\n\n", header)
}

func (c Console) PrintRegular(format string, args ...any) {
	fmt.Printf(format, args...)
	fmt.Println()
}

func (c Console) PrintNewLine() {
	fmt.Println()
}

func (c Console) Prompt(text string) string {
	fmt.Printf("%v: ", text)

	var result string
	_, _ = fmt.Scanf("%s", &result)

	return result
}

func (c Console) PromptPassword(text string) string {
	fmt.Printf("%v: ", text)

	password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return ""
	}
	fmt.Println()
	return string(password)
}
