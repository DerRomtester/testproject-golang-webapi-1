build:
	go build -o bin/main cmd/main.go

run:
	go run cmd/main.go

clean: 
	rm bin/main
	go clean

start-container: 
	docker-compose build
	docker compose up