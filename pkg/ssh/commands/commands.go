package commands

import (
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

var CommandHandlers = map[string]func(connection *ssh.ServerConn, payload []byte) (string, error){}

func RegisterCommand(name string, handler func(connection *ssh.ServerConn, payload []byte) (string, error)) {
	CommandHandlers[name] = handler
}

func HandlePwd(connection *ssh.ServerConn, payload []byte) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir + "\n", nil
}

func HandleCat(connection *ssh.ServerConn, payload []byte) (string, error) {
	words := strings.Split(string(payload), " ")
	if len(words) < 1 {
		return "Usage: cat <file>\n", nil
	}
	fileName := words[0]
	file, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(file) + "\n", nil
}

func HandleListFiles(connection *ssh.ServerConn, payload []byte) (string, error) {
	files, err := os.ReadDir(".")
	if err != nil {
		return "", err
	}
	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return strings.Join(fileNames, "\n") + "\n", nil
}

func HandleEcho(connection *ssh.ServerConn, payload []byte) (string, error) {
	return "You echoed: " + string(payload) + "\n", nil
}

func HandleClear(connection *ssh.ServerConn, payload []byte) (string, error) {
	return "\033[H\033[2J", nil
}

func HandleExit(connection *ssh.ServerConn, payload []byte) (string, error) {
	return "Bye\n", nil
}

func HandleHelp(connection *ssh.ServerConn, payload []byte) (string, error) {
	commands := []string{}
	for cmd := range CommandHandlers {
		commands = append(commands, cmd)
	}
	commandList := "Available commands:\n" + strings.Join(commands, "\n") + "\n"
	return commandList, nil
}

func HandleWhoami(connection *ssh.ServerConn, payload []byte) (string, error) {
	if connection == nil {
		return "You are anonymous\n", nil
	}
	username := connection.User()
	if username == "" {
		username = "anonymous"
	}
	return "You are " + username + "\n", nil
}
