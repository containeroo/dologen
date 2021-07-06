package main

import (
	"encoding/base64"
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
)

func main() {
	username := flag.StringP("username", "u", "", "username for docker registry")
	password := flag.StringP("password", "p", "", "password for docker registry")
	server := flag.StringP("server", "s", "", "docker registry server")
	base64Output := flag.BoolP("base64", "b", false, "output result base64 encoded")
	flag.Parse()

	if *username == "" {
		fmt.Println("username cannot be empty!")
		flag.Usage()
		os.Exit(1)
	}
	if *password == "" {
		fmt.Println("password cannot be empty!")
		flag.Usage()
		os.Exit(1)
	}
	if *server == "" {
		fmt.Println("server cannot be empty!")
		flag.Usage()
		os.Exit(1)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", *username, *password)))

	result := fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\",\"auth\":\"%s\"}}}", *server, *username, *password, auth)

	if *base64Output {
		fmt.Println(base64.StdEncoding.EncodeToString([]byte(result)))
	} else {
		fmt.Println(result)
	}
}
