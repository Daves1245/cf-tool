package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/mitchellh/go-homedir"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/cmd"
	"github.com/xalanq/cf-tool/config"
	docopt "github.com/docopt/docopt-go"
)

const version = "v1.0.0"
const configPath = "~/.cf/config"
const sessionPath = "~/.cf/session"

func help() {
	fmt.Println("Usage:")
	fmt.Println("  cf config")
	fmt.Println("  cf submit [-f <file>] [<specifier>...]")
	fmt.Println("  cf list [<specifier>...]")
	fmt.Println("  cf parse [<specifier>...]")
	fmt.Println("  cf gen [<alias>]")
	fmt.Println("  cf test [<file>]")
	fmt.Println("  cf watch [all] [<specifier>...]")
	fmt.Println("  cf open [<specifier>...]")
	fmt.Println("  cf stand [<specifier>...]")
	fmt.Println("  cf sid [<specifier>...]")
	fmt.Println("  cf race [<specifier>...]")
	fmt.Println("  cf pull [ac] [<specifier>...]")
	fmt.Println("  cf clone [ac] [<handle>]")
	fmt.Println("  cf upgrade")
}

func main() {
	color.Output = ansi.NewAnsiStdout()

	cfgPath, err := homedir.Expand(configPath)
	if err != nil {
		color.Red("Failed to expand config path: %v", err)
		return
	}
	color.Yellow("Config path: %s", cfgPath)

	clnPath, err := homedir.Expand(sessionPath)
	if err != nil {
		color.Red("Failed to expand session path: %v", err)
		return
	}
	color.Yellow("Session path: %s", clnPath)

	config.Init(cfgPath)
	if config.Instance == nil {
		color.Red("Failed to initialize config")
		return
	}
	color.Yellow("Config initialized with host: %s", config.Instance.Host)

	client.Init(clnPath, config.Instance.Host, config.Instance.Proxy)
	if client.Instance == nil {
		color.Red("Failed to initialize client")
		return
	}
	color.Yellow("Client initialized")

	usage := `Codeforces Tool $%version%$ (cf). https://github.com/xalanq/cf-tool

You should run "cf config" to configure your handle, password and code
templates at first.

If you want to compete, the best command is "cf race"

Usage:
  cf config
  cf submit [-f <file>] [<specifier>...]
  cf list [<specifier>...]
  cf parse [<specifier>...]
  cf gen [<alias>]
  cf test [<file>]
  cf watch [all] [<specifier>...]
  cf open [<specifier>...]
  cf stand [<specifier>...]
  cf sid [<specifier>...]
  cf race [<specifier>...]
  cf pull [ac] [<specifier>...]
  cf clone [ac] [<handle>]
  cf upgrade

Options:
  -h --help            Show this screen.
  --version            Show version.
  -f <file>, --file <file>, <file>
                       Path to file. E.g. "a.cpp", "./temp/a.cpp"
  <specifier>          Any useful text. E.g.
                       "https://codeforces.com/contest/100",
                       "https://codeforces.com/contest/180/problem/A",
                       "https://codeforces.com/group/Cw4JRyRGXR/contest/269760"
                       "1111A", "1111", "a", "Cw4JRyRGXR"
                       You can combine multiple specifiers to specify what you
                       want.
  <alias>              Template's alias. E.g. "cpp"
  ac                   The status of the submission is Accepted.`

	usage = strings.Replace(usage, `$%version%$`, version, 1)
	opts, _ := docopt.ParseArgs(usage, os.Args[1:], fmt.Sprintf("Codeforces Tool (cf) %v", version))
	opts[`{version}`] = version

	err = cmd.Eval(opts)
	if err != nil {
		color.Red(err.Error())
	}
	color.Unset()
}
