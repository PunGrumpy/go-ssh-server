<div align="center">
    <h1><code>🐰</code> SSH Server</h1>
    <strong>SSH Server</strong> is a simple SSH server written in Go.
    <div>
        <img src="./.github/images/Draugr.gif" width="521" />
    </div>
</div>

## `📸` Screenshots

![Screenshot](.github/images/screenshot.png)

## `📝` About

This project is a simple SSH server written in Go. It is intended to be used as a library for other projects, but it can also be used as a standalone SSH server.

## `🚀` Usage

### `🔐` Key Generation

```bash
go run cmd/keygen/main.go
```

```bash
make keygen
```

### `🏭` Server

**[Server](cmd/server/main.go)**

```bash
go run cmd/server/main.go
```

```bash
make server
```

### `📦` Client

- **[Client](cmd/client/main.go)** Execute a command

```bash
go run cmd/client/main.go
```

```bash
ssh localhost -p 2022 -i server_key.pem "whoami"
```

```bash
make exec
```

- Interactive shell

```bash
ssh localhost -p 2023 -i server_key.pem
```

```bash
make shell
```

## `🐳` Usage on Docker

📌 Don't forget use `keygen` before run docker

```shell
docker compose build # docker-compose build
docker compose up -d # docker-compose up -d
```

## `📃` Available Commands

- `help` - Show help
- `exit` - Exit shell
- `whoami` - Show current user _still not working_
- `ls` - List files
- `pwd` - Show current directory
- `cat` - Show file content
- `echo` - Print text
- `clear` - Clear screen

## `📚` References

- [gliderlabs/ssh](https://pkg.go.dev/github.com/gliderlabs/ssh?utm_source=godoc) _Learned a lot from this project_
- [golang/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh?utm_source=godoc)
- [golang/crypto](https://pkg.go.dev/golang.org/x/crypto?utm_source=godoc)
- [golang/term](https://pkg.go.dev/golang.org/x/term?utm_source=godoc)
- [golang/net](https://pkg.go.dev/golang.org/x/net?utm_source=godoc)
- [golang/sys](https://pkg.go.dev/golang.org/x/sys?utm_source=godoc)

## `⚖️` License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
