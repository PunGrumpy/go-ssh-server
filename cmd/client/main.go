package main

import (
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {
	var (
		privateKey []byte
		publicKey  []byte
		err        error
	)

	privateKey, err = os.ReadFile("server_key.pem")
	if err != nil {
		log.Fatalf("unable to read server key: %s", err.Error())
	}

	publicKey, err = os.ReadFile("server_key.pub")
	if err != nil {
		log.Fatalf("unable to read authorized key: %s", err.Error())
	}

	privateKeyParsed, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		log.Fatalf("unable to parse private key: %s", err.Error())
	}

	publicKeyParsed, _, _, _, err := ssh.ParseAuthorizedKey(publicKey)
	if err != nil {
		log.Fatalf("unable to parse public key: %s", err.Error())
	}

	config := &ssh.ClientConfig{
		User: "kmitl",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKeyParsed),
		},
		HostKeyCallback: ssh.FixedHostKey(publicKeyParsed),
	}

	client, err := ssh.Dial("tcp", "localhost:2023", config)
	if err != nil {
		log.Fatalf("unable to connect: %s", err.Error())
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("unable to create session: %s", err.Error())
	}

	defer session.Close()

	output, err := session.Output("whoami")
	if err != nil {
		log.Fatalf("session output error: %s", err.Error())
	}

	log.Printf("Output from remote host: %s", string(output))
}
