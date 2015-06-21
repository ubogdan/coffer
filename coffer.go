package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/howeyc/gopass"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

const emptyCoffer string = `{"secrets":{}}`

type KeybaseStatus struct {
	Status struct {
		Configured bool `json:"configured"`
		LoggedIn   bool `json:"logged_in"`
	} `json:"status"`
	User struct {
		Name string `json:"name"`
		Key  struct {
			KeyID       string `json:"key_id"`
			Fingerprint string `json:"fingerprint"`
		} `json:"key"`
		Proofs struct {
			Twitter        string   `json:"twitter"`
			Github         string   `json:"github"`
			Reddit         string   `json:"reddit"`
			Hackernews     string   `json:"hackernews"`
			GenericWebSite []string `json:"generic_web_site"`
			DNS            []string `json:"dns"`
		} `json:"proofs"`
	} `json:"user"`
}

type Coffer struct {
	Secrets map[string]string `json:"secrets"`
}

func main() {
	app := cli.NewApp()
	app.Usage = "A secret storage utility, built on keybase and PGP"
	app.Version = "0.0.0"

	app.Commands = []cli.Command{
		{
			Name:    "store",
			Aliases: []string{"s"},
			Usage:   "store a secret",
			Action: func(c *cli.Context) {
				name := c.Args().First()
				fmt.Printf("Type the new secret:")
				secret := gopass.GetPasswd()
				fmt.Printf("Type again to confirm:")
				secretConfirm := gopass.GetPasswd()
				if string(secretConfirm) != string(secret) {
					fmt.Println("Secrets do not match!")
					return
				}

				if string(secret) == "" {
					fmt.Println("The empty string is not a valid secret")
					return
				}

				coffer := readEncryptedFile()
				coffer.Secrets[name] = string(secret)
				coffer.writeEncryptedFile()
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l", "ls"},
			Usage:   "list the currently stored secrets",
			Action: func(c *cli.Context) {
				coffer := readEncryptedFile()

				fmt.Println("The currently stored secrets are:")
				for k := range coffer.Secrets {
					fmt.Println(k)
				}
			},
		},
		{
			Name:    "get",
			Aliases: []string{"g"},
			Usage:   "get a secret that matches the given key (if any)",
			Action: func(c *cli.Context) {
				coffer := readEncryptedFile()
				name := c.Args().First()
				secret := coffer.Secrets[name]
				if secret == "" {
					fmt.Printf("No secret with name `%s`\n", name)
				} else {
					fmt.Println(secret)
				}
			},
		},
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "create an empty coffer",
			Action: func(c *cli.Context) {
				fmt.Printf("This will\033[1m destroy\033[0m any existing coffer. Continue? [y/N] ")
				var input string
				fmt.Scanln(&input)
				if !strings.EqualFold(input, "y") {
					fmt.Println("Exiting, the existing coffer has not been modified.")
					return
				}

				username := getKeybaseUsername()
				filepath := getCofferFilePath()
				keybase := exec.Command("keybase", "encrypt", username, "--message", emptyCoffer, "--output", filepath)
				if err := keybase.Run(); err != nil {
					panic(err)
				}
				fmt.Printf("Empty coffer created at '%s'\n", filepath)
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "delete a secret from the coffer",
			Action: func(c *cli.Context) {
				name := c.Args().First()
				fmt.Printf("This will\033[1m delete\033[0m the secret for `%s`. Continue? [y/N] ", name)
				var input string
				fmt.Scanln(&input)
				if !strings.EqualFold(input, "y") {
					fmt.Println("Exiting, the coffer has not been modified.")
					return
				}

				coffer := readEncryptedFile()
				delete(coffer.Secrets, name)
				coffer.writeEncryptedFile()
			},
		},
	}

	app.Run(os.Args)
}

func getCofferFilePath() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(usr.HomeDir, "/.coffer")
}

func (c Coffer) writeEncryptedFile() {
	jsonStr, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		return
	}
	username := getKeybaseUsername()
	filepath := getCofferFilePath()
	keybase := exec.Command("keybase", "encrypt", username, "--message", string(jsonStr), "--output", filepath)
	if err := keybase.Run(); err != nil {
		panic(err)
	}
}

func cleanupAndExit(status int, paths ...string) {
	for _, p := range paths {
		os.RemoveAll(p)
	}
	os.Exit(status)
}

func readEncryptedFile() Coffer {

	dir, err := ioutil.TempDir("", "coffer")
	defer os.RemoveAll(dir)
	if err != nil {
		fmt.Println("Failed to create temporary directory")
		cleanupAndExit(1, dir)
	}

	file, err := ioutil.TempFile(dir, "coffer")
	if err != nil {
		fmt.Println("Failed to create temporary file")
		cleanupAndExit(1, dir)
	}

	keybase := exec.Command("keybase", "decrypt", getCofferFilePath(), "--output", file.Name())

	if err := keybase.Run(); err != nil {
		fmt.Println("Keybase failed to decrypt the coffer!")
		fmt.Println("If you have not yet created a coffer, please run `coffer create`.")
		cleanupAndExit(1, dir)
	}

	content, err := ioutil.ReadFile(file.Name())
	if err != nil {
		fmt.Println("Failed to read temporary coffer file")
		cleanupAndExit(1, dir)
	}

	var coffer Coffer
	if err := json.Unmarshal(content, &coffer); err != nil {
		fmt.Println("Failed to unmarshal JSON to struct")
		cleanupAndExit(1, dir)
	}

	return coffer
}

func getKeybaseUsername() string {
	keybase := exec.Command("keybase", "status")
	var out bytes.Buffer

	keybase.Stdout = &out
	err := keybase.Run()
	if err != nil {
		log.Fatal(err)
	}

	res := &KeybaseStatus{}
	err = json.Unmarshal(out.Bytes(), &res)
	if err != nil {
		panic(err)
	}

	return res.User.Name
}
