package ssh

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/PunGrumpy/go-ssh/pkg/ssh/commands"
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

func ParseCommandArgs(payload string) (string, string) {
	parts := strings.SplitN(payload, " ", 2)
	command := parts[0]
	var args string
	if len(parts) > 1 {
		args = parts[1]
	}
	return command, args
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
			if err := newChannel.Reject(ssh.UnknownChannelType, "unknown channel type"); err != nil {
				log.Fatalf("unable to reject channel: %v", err)
			}
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
			output := ExecCommand([]byte(payload))
			if _, err := channel.Write([]byte(output)); err != nil {
				log.Fatalf("could not write output: %v", err)
			}
			exitStatus := []byte{0, 0, 0, 0}
			if _, err := channel.SendRequest("exit-status", false, exitStatus); err != nil {
				log.Fatalf("could not send exit status: %v", err)
			}
			if err := req.Reply(true, nil); err != nil {
				log.Fatalf("could not reply to request: %v", err)
			}
			channel.Close()
		case "shell": // Start an interactive shell
			if err := req.Reply(req.Type == "shell", nil); err != nil {
				log.Fatalf("could not reply to request: %v", err)
			}
		case "pty-req": // Request a pseudo terminal
			CreateTerminal(nil, channel)
			if err := req.Reply(true, nil); err != nil {
				log.Fatalf("could not reply to request: %v", err)
			}
		case "env": // Set environment variables
			if err := req.Reply(true, nil); err != nil {
				log.Fatalf("could not reply to request: %v", err)
			}
		case "subsystem": // Start a subsystem
			subsystem := string(req.Payload[4:])
			switch subsystem {
			case "sftp":
				HandleDataTransfer(channel, req, "SFTP")
			case "scp":
				HandleDataTransfer(channel, req, "SCP")
			default:
				if err := req.Reply(false, nil); err != nil {
					log.Fatalf("could not reply to request: %v", err)
				}
			}
		default:
			if err := req.Reply(false, nil); err != nil {
				log.Fatalf("could not reply to request: %v", err)
			}
		}
	}
}

func HandleDataTransfer(channel ssh.Channel, req *ssh.Request, name string) {
	log.Printf("Starting %s server...\n", name)
	if err := req.Reply(true, nil); err != nil {
		log.Fatalf("could not reply to request: %v", err)
	}

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

func ExecCommand(payload []byte) string {
	command, args := ParseCommandArgs(string(payload))
	handler, ok := commands.CommandHandlers[command]
	if !ok {
		return "Unknown command\n"
	}

	result, err := handler(nil, []byte(args))
	if err != nil {
		return err.Error()
	}
	return result
}

func CreateTerminal(connection *ssh.ServerConn, channel ssh.Channel) {
	terminalInstance := term.NewTerminal(channel, "› ")

	go func() {
		defer channel.Close()
		if _, err := terminalInstance.Write([]byte("Welcome to the SSH server\n")); err != nil {
			log.Fatalf("unable to write to terminal: %v", err)
		}
		if _, err := terminalInstance.Write([]byte("Type 'exit' to close the connection\n")); err != nil {
			log.Fatalf("unable to write to terminal: %v", err)
		}
		if _, err := terminalInstance.Write([]byte("Type 'help' to see all available commands\n")); err != nil {
			log.Fatalf("unable to write to terminal: %v", err)
		}

		for {
			line, err := terminalInstance.ReadLine()
			if err != nil {
				fmt.Printf("unable to read line: %v", err)
				return
			}

			command, args := ParseCommandArgs(line)
			handler, ok := commands.CommandHandlers[command]
			if !ok {
				if _, err := terminalInstance.Write([]byte("Unknown command\n")); err != nil {
					log.Fatalf("unable to write to terminal: %v", err)
				}
				continue
			}

			result, err := handler(connection, []byte(args))
			if err != nil {
				if _, err := terminalInstance.Write([]byte(err.Error())); err != nil {
					log.Fatalf("unable to write to terminal: %v", err)
				}
				continue
			}
			if _, err := terminalInstance.Write([]byte(result)); err != nil {
				log.Fatalf("unable to write to terminal: %v", err)
			}
		}
	}()
}
