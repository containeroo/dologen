package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	flag "github.com/spf13/pflag"
	"io/ioutil"
	"os"
)

const version = "1.1.0"

func main() {
	username := flag.StringP("username", "u", "", "username for docker registry")
	password := flag.StringP("password", "p", "", "password for docker registry")
	passwordFile := flag.StringP("password-file", "f", "", "password file for docker registry")
	server := flag.StringP("server", "s", "", "docker registry server")
	base64Output := flag.BoolP("base64", "b", false, "output result base64 encoded")
	printVersion := flag.BoolP("version", "v", false, "Print the current version and exit")
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *username == "" {
		fmt.Println("username cannot be empty!")
		flag.Usage()
		os.Exit(1)
	}
	if *password == "" && *passwordFile == "" {
		fmt.Println("password cannot be empty!")
		flag.Usage()
		os.Exit(1)
	}
	if *server == "" {
		fmt.Println("server cannot be empty!")
		flag.Usage()
		os.Exit(1)
	}

	if *passwordFile != "" {
		passwordBytes, err := ioutil.ReadFile(*passwordFile)
		if err != nil {
			fmt.Println("read password file error:", err)
			os.Exit(1)
		}

		passwordBytes = bytes.Replace(passwordBytes, []byte("\n"), []byte(""), -1)
		*password = string(passwordBytes)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", *username, *password)))

	result := fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\",\"auth\":\"%s\"}}}", *server, *username, *password, auth)

	switch *base64Output {
	case true:
		fmt.Println(base64.StdEncoding.EncodeToString([]byte(result)))
	default:
		fmt.Println(result)
	}
}
