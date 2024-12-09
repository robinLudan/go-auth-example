project = go-auth-example

run:
	@go run cmd/${project}/main.go

build:
	@go build -o bin/${project} cmd/${project}/main.go

test:
	@go test -v ./...

clean:
	@rm -rf bin/${project}
