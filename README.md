# Go Auth Example

This is aimed to be a reference implementation of authentication service in go.

## Features

- User registration
- User login
- Get currently authenticated user
- [JWT Authentication](https://pkg.go.dev/github.com/golang-jwt/jwt/v5)
- [Graceful shutdown](https://pkg.go.dev/net/http#Server.Shutdown)

## Installation and configuration

- Clone the repo

  ```bash
  git clone https://github.com/robinLudan/go-auth-example.git
  cd go-auth-example
  ```

- Set env variables

  ```bash
  cp .env.example .env
  ```

## Usage

### Build and run the service

Default

```bash
go build -o bin/go-auth-example cmd/go-auth-example/main.go
./bin/go-auth-example
```

Or using make

```bash
make build
./bin/go-auth-example
```

### API endpoints

- POST `/register` - Register a new user
- POST `/login` - Login a user
- GET `/me` - Get currently authenticated user
- GET `/errors` - What to expect as errors

## Contributing

Open an issue or submit a PR for any improvements or bug fixes.
