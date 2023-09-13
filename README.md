<div align="center">
    <h1><code>🐰</code> SSH Server</h1>
    <strong>SSH Server</strong> is a simple SSH server written in Go.
</div>

## `📝` About

This project is a simple SSH server written in Go. It is intended to be used as a library for other projects, but it can also be used as a standalone SSH server.

## `🚀` Usage

### `🔐` Key Generation

```bash
go run cmd/keygen/main.go
```

### `🏭` Server

**[Server](cmd/server/main.go)**

```bash
go run cmd/server/main.go
```

### `📦` Client

- **[Client](cmd/client/main.go)** Execute a command

```bash
go run cmd/client/main.go
```

```bash
ssh localhost -p 2022 -i server_key.pem "whoami"
```

- Interactive shell

```bash
ssh localhost -p 2023 -i server_key.pem
```

## `⚖️` License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
