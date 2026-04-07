# Go TCP Chat

Minimal multi-client chat server written in Go.

## Run

```bash
go run main.go
```

Server listens on `localhost:7452`.

## Usage

Connect from terminal clients (example with `nc`):

```bash
nc localhost 7452
```

Message format:

- `/username:message`
- `/annonym:message`
- `/quit`

## Demo Screenshot

![Chat demo](Screenshot%20from%202026-04-07%2018-08-21.png)
