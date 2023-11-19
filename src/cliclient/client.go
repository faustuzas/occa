package cliclient

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"go.uber.org/zap"
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
	logger, err := zap.NewDevelopmentConfig().Build(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		panic(err)
	}

	gatewayClient := gatewayclient.New(params.GatewayAddress, logger)

	app := cliClientApplication{
		Params:  params,
		gateway: gatewayClient,
		console: Console{},
	}

	app.preExistHooks = append(app.preExistHooks, gatewayClient.Close)

	steps := []func() bool{
		app.printIntro,
		app.checkGatewayConnectivity,
		app.authentication,
		app.mainMenu,
	}

	for _, stage := range steps {
		if !stage() {
			return
		}
	}

	for _, h := range app.preExistHooks {
		h()
	}
}

type cliClientApplication struct {
	Params

	gateway *gatewayclient.Client
	console Console

	preExistHooks []func()
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

func (a cliClientApplication) mainMenu() bool {
	a.printMenu()

	for {
		a.console.PrintNewLine()
		res := a.console.Prompt("Select the option by entering the number and hitting Enter")
		option, err := strconv.Atoi(res)
		if err != nil {
			a.console.PrintRegular("Failed to parse the option. Reason: %v", err)
			continue
		}

		var action func()
		switch option {
		case 1:
			action = a.printActiveUsers
		case 2:
			action = a.printMenu
		case 3:
			a.console.PrintNewLine()
			a.console.PrintRegular("Good bye!")
			return false
		default:
			a.console.PrintRegular("Option %d not recognized", option)
			continue
		}

		action()
	}
}

func (a cliClientApplication) printMenu() {
	a.console.ClearScreen()

	a.console.PrintHeader("MENU")
	a.console.PrintRegular("1. Print active users")
	a.console.PrintRegular("2. Clear screen")
	a.console.PrintRegular("3. Exit")
}

func (a cliClientApplication) printActiveUsers() {
	users, err := a.gateway.ActiveUsers(a.ctx())
	if err != nil {
		a.console.PrintRegular("Failed to fetch active users. Reason: %v", err)
		return
	}

	a.console.PrintRegular("Active users:")
	for _, u := range users {
		a.console.PrintRegular("* %s", u)
	}
	a.console.PrintNewLine()
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

func (c Console) ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
