package main

import (
	"log"
	"os"

	"github.com/PunGrumpy/go-ssh/pkg/ssh"
)

func main() {
	var (
		serverKeyBytes     []byte
		authorizedKeyBytes []byte
		err                error
	)

	serverKeyBytes, err = os.ReadFile("server_key.pem")
	if err != nil {
		log.Fatalf("unable to read server key: %s", err.Error())
	}

	authorizedKeyBytes, err = os.ReadFile("server_key.pub")
	if err != nil {
		log.Fatalf("unable to read authorized key: %s", err.Error())
	}

	if err = ssh.StartServer(serverKeyBytes, authorizedKeyBytes); err != nil {
		log.Fatalf("unable to start server: %s", err.Error())
	}
}
