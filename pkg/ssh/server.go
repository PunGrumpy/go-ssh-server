package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func StartServer(privateKey []byte, authorizedKey []byte) error {
	authorizedKeysMap, err := ParseAuthorizedKey(authorizedKey)
	if err != nil {
		return err
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(context ssh.ConnMetadata, publicKey ssh.PublicKey) (*ssh.Permissions, error) {
			return PublicKeyCallback(context, publicKey, authorizedKeysMap)
		},
	}

	private, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return errors.New("unable to parse private key: " + err.Error())
	}

	config.AddHostKey(private)

	log.Println("Starting server on port 2023...")
	listener, err := net.Listen("tcp", "0.0.0.0:2023")
	if err != nil {
		return errors.New("unable to start server: " + err.Error())
	}

	for {
		netConnection, err := listener.Accept()
		if err != nil {
			log.Println("unable to accept connection: " + err.Error())
			continue
		}

		connection, channels, requests, err := ssh.NewServerConn(netConnection, config)
		if err != nil {
			log.Println("unable to handshake: " + err.Error())
			continue
		}
		if connection != nil && connection.Permissions != nil {
			log.Printf(
				"Logged in with key: %s for %s",
				connection.Permissions.Extensions["pubkey-fp"],
				connection.User(),
			)
		}

		go ssh.DiscardRequests(requests)
		go HandleConnection(connection, channels)
	}
}

func ParseAuthorizedKey(authorizedKey []byte) (map[string]bool, error) {
	authorizedKeysMap := map[string]bool{}
	for len(authorizedKey) > 0 {
		publicKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKey)
		if err != nil {
			return nil, errors.New("unable to parse authorized key: " + err.Error())
		}

		authorizedKeysMap[string(publicKey.Marshal())] = true
		authorizedKey = rest
	}
	return authorizedKeysMap, nil
}

func PublicKeyCallback(context ssh.ConnMetadata, publicKey ssh.PublicKey, authorizedKeysMap map[string]bool) (*ssh.Permissions, error) {
	if authorizedKeysMap[string(publicKey.Marshal())] {
		return &ssh.Permissions{
			Extensions: map[string]string{
				"pubkey-fp": ssh.FingerprintSHA256(publicKey),
			},
		}, nil
	}

	return nil, errors.New("unknown public key for " + context.User())
}

func HandleConnection(connection *ssh.ServerConn, channels <-chan ssh.NewChannel) {
	for newChannel := range channels {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Println("could not accept channel (" + err.Error() + ")")
			continue
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				log.Printf("request type made by client: %s\n", req.Type)
				switch req.Type {
				case "exec": // Execute a command
					payload := bytes.TrimPrefix(req.Payload, []byte{0, 0, 0, 6}) // 0 0 0 6 is the length of "exec"
					channel.Write([]byte(ExecCommand(connection, payload)))
					channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0}) // 0 is the exit status (success)
					req.Reply(true, nil)
					channel.Close()
				case "shell": // Start an interactive shell
					req.Reply(req.Type == "shell", nil)
				case "pty-req": // Request a pseudo terminal
					CreateTerminal(connection, channel)
					req.Reply(true, nil)
				default:
					req.Reply(false, nil)
				}
			}
		}(requests)
	}
}

func CreateTerminal(connection *ssh.ServerConn, channel ssh.Channel) {
	terminalInstance := term.NewTerminal(channel, "› ")

	go func() {
		defer channel.Close()
		terminalInstance.Write([]byte("Welcome to the SSH server\n"))
		terminalInstance.Write([]byte("Type 'exit' to close the connection\n"))
		terminalInstance.Write([]byte("Type 'help' to see all available commands\n"))

		for {
			line, err := terminalInstance.ReadLine()
			if err != nil {
				fmt.Printf("unable to read line: %v", err)
				return
			}

			switch line {
			case "whoami":
				terminalInstance.Write([]byte(ExecCommand(connection, []byte("whoami"))))
			case "help":
				terminalInstance.Write([]byte(ExecCommand(connection, []byte("help"))))
			case "exit":
				terminalInstance.Write([]byte(ExecCommand(connection, []byte("exit"))))
				return
			default:
				terminalInstance.Write([]byte("Unknown command\n"))
			}
		}
	}()
}

func ExecCommand(connection *ssh.ServerConn, payload []byte) string {
	switch string(payload) {
	case "whoami":
		return fmt.Sprintf("You are %s\n", connection.User())
	case "exit":
		return "Bye\n"
	case "help":
		return "Available commands:\nwhoami\nexit\n"
	default:
		return fmt.Sprintf("Unknown command: %s\n", string(payload))
	}
}
