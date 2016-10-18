package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/lunixbochs/go-keychain"
	"github.com/urfave/cli"
	"github.com/howeyc/gopass"
	"github.com/ghodss/yaml"
)

func main() {
	app := cli.NewApp()
	app.Name = "pepper"
	app.Version = "0.1.1"
	app.Usage = "pepper <target> <function> [ARGUMENTS ...]"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "hostname, H",
			Usage:  "Salt API hostname. Should include http[s]//.",
			EnvVar: "SALT_HOST",
		},
		cli.StringFlag{
			Name:   "username, u",
			Usage:  "Salt API username.",
			EnvVar: "SALT_USER",
		},
		cli.StringFlag{
			Name:   "password, p",
			Usage:  "Salt API password.",
			EnvVar: "SALT_PASSWORD",
		},
		cli.BoolFlag{
			Name:   "stdin, P",
			Usage:  "Salt API password. Gets from stdin.",
		},
		cli.StringFlag{
			Name:   "auth, a",
			Value:  "pam",
			Usage:  "Salt authentication method.",
			EnvVar: "SALT_AUTH",
		},
		cli.BoolFlag{
			Name:   "yaml, Y",
			Usage:  "Output as YAML.",
			EnvVar: "SALT_YAML",
		},
	}
	app.Action = func(c *cli.Context) error {
		if len(c.Args()) < 2 {
			fmt.Println("pepper <target> <function> [ARGUMENTS ...]")
			return nil
		}

		hostname := c.String("hostname")
		username := c.String("username")
		password := c.String("password")
		auth := c.String("auth")

		// Read password from stdin if flagged
		if c.Bool("stdin") {
			fmt.Printf("Password: ")
			pwd, err := gopass.GetPasswd()
			if err != nil {
				log.Fatal(err)
			}
			password = string(pwd)
		}

		// Try to save the password to the macOS keychain 
		// if it was read from stdin
		if c.Bool("stdin") == true && runtime.GOOS == "darwin" {
			keychain.Remove("pepper", username)
			keychain.Add("pepper", username, password)
		}

		// Try getting the password from the macOS Keychain
		var keychainpwd string
		if runtime.GOOS == "darwin" && c.Bool("stdin") == false {
   			keychainpwd, err := keychain.Find("pepper", username)
   			if err == nil {
   				password = keychainpwd
   			}
		}

		salt := NewSalt(hostname)

		err := salt.Login(username, password, auth)
		if err != nil {
			if keychainpwd != "" {
				log.Fatal("Using macOS keychain\n", err)
			} else {
				log.Fatal(err)
			}
		}

		target := c.Args().Get(0)
		function := c.Args().Get(1)
		arguments := c.Args().Get(2)

		response, _ := salt.Run(target, function, arguments)

		// Convert output to YAML if flagged
		var output string
		if c.Bool("yaml") {
			y, err := yaml.JSONToYAML(response)
			output = string(y)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			output = string(response)
		}

		fmt.Println(output)
		return nil
	}

	app.Run(os.Args)

}
