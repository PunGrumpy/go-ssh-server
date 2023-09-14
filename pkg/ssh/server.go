package ssh

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

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

		go HandleSession(channel, requests)
	}
}

func HandleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		log.Printf("request type made by client: %s\n", req.Type)
		switch req.Type {
		case "exec": // Execute a command
			payload := string(req.Payload[4:]) // Make sure to remove the length of the payload
			output := ExecCommand(nil, []byte(payload))
			channel.Write([]byte(output))
			exitStatus := []byte{0, 0, 0, 0}
			channel.SendRequest("exit-status", false, exitStatus)
			req.Reply(true, nil)
			channel.Close()
		case "shell": // Start an interactive shell
			req.Reply(req.Type == "shell", nil)
		case "pty-req": // Request a pseudo terminal
			CreateTerminal(nil, channel)
			req.Reply(true, nil)
		case "env": // Set environment variables
			req.Reply(true, nil)
		case "subsystem": // Start a subsystem
			subsystem := string(req.Payload[4:])
			switch subsystem {
			case "sftp":
				HandleDataTransfer(channel, req, "SFTP")
			case "scp":
				HandleDataTransfer(channel, req, "SCP")
			default:
				req.Reply(false, nil)
			}
		default:
			req.Reply(false, nil)
		}
	}

}

func HandleDataTransfer(channel ssh.Channel, req *ssh.Request, name string) {
	log.Printf("Starting %s server...\n", name)
	req.Reply(true, nil)

	serverOptions := []sftp.ServerOption{
		sftp.WithDebug(os.Stdout),
	}

	server, err := sftp.NewServer(
		channel,
		serverOptions...,
	)
	if err != nil {
		log.Fatalf("unable to start %s server: %v", name, err)
	}

	if err := server.Serve(); err != nil {
		log.Fatalf("unable to start %s server: %v", name, err)
	}
}

func ExecCommand(connection *ssh.ServerConn, payload []byte) string {
	switch string(payload) {
	case "whoami":
		return fmt.Sprintf("You are %s\n", connection.User())
	case "ls":
		files, err := os.ReadDir(".")
		if err != nil {
			log.Fatalf("unable to read directory: %v", err)
		}
		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, file.Name())
		}
		return strings.Join(fileNames, "\n") + "\n"
	case "echo":
		return "Enter text to echo: "
	case "clear":
		return "\033[H\033[2J"
	case "exit":
		return "Bye\n"
	case "help":
		return GetAvailableCommands()
	default:
		return fmt.Sprintf("Unknown command: %s\n", string(payload))
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
				terminalInstance.Write([]byte("You are " + connection.User() + "\n"))
			case "ls":
				files, err := os.ReadDir(".")
				if err != nil {
					log.Fatalf("unable to read directory: %v", err)
				}
				for _, file := range files {
					terminalInstance.Write([]byte(file.Name() + "\n"))
				}
			case "echo":
				terminalInstance.Write([]byte("Enter text to echo: "))
				text, err := terminalInstance.ReadLine()
				if err != nil {
					fmt.Printf("unable to read line: %v", err)
					return
				}
				terminalInstance.Write([]byte("You echoed: " + text + "\n"))
			case "clear":
				terminalInstance.Write([]byte("\033[H\033[2J"))
			case "help":
				terminalInstance.Write([]byte(GetAvailableCommands()))
			case "exit":
				terminalInstance.Write([]byte("Bye\n"))
				return
			default:
				terminalInstance.Write([]byte("Unknown command\n"))
			}
		}
	}()
}

func GetAvailableCommands() string {
	commands := []string{"whoami", "echo", "clear", "exit", "help"}
	commandList := "Available commands:\n" + strings.Join(commands, "\n") + "\n"
	return commandList
}
