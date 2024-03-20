build:
	@docker-compose up --build -d

down:
	@docker-compose down

test:
	@docker-compose exec app go test -v ./store