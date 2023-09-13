<div align="center">
    <h1><code>🐰</code> SSH Server</h1>
    <strong>SSH Server</strong> is a simple SSH server written in Go.
</div>

## `📝` About

This project is a simple SSH server written in Go. It is intended to be used as a library for other projects, but it can also be used as a standalone SSH server.

## `🚀` Usage

### `🏭` Server

**[Server](cmd/server/main.go)**

```go
go run cmd/server/main.go
```

### `📦` Client

**[Client](cmd/client/main.go)**

- **[Client](cmd/client/main.go)** Execute a command

```go
go run cmd/client/main.go
```

- Interactive shell

```go
ssh localhost -p 2023 -i server_key.pem
```

## `⚖️` License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
