package main

import (
	"log"
	"os"

	"github.com/PunGrumpy/go-ssh/pkg/ssh"
)

func main() {
	const (
		READABLE_FILE_MODE = 0644
		WRITABLE_FILE_MODE = 0600
	)

	var (
		privateKey, publicKey []byte
		err                   error
	)

	if privateKey, publicKey, err = ssh.GenerateKey(); err != nil {
		log.Fatalf("unable to generate key: %s", err.Error())
	}

	if err = os.WriteFile("server_key.pem", privateKey, WRITABLE_FILE_MODE); err != nil {
		log.Fatalf("unable to write private key to file: %s", err.Error())
	}

	if err = os.WriteFile("server_key.pub", publicKey, READABLE_FILE_MODE); err != nil {
		log.Fatalf("unable to write public key to file: %s", err.Error())
	}
}
