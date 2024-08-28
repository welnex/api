coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

swag:
	swag init -g ./cmd/api/main.go

clean:
	rm coverage.out
	rm coverage.html
