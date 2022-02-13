help:
	@echo "build      - Build database image."
	@echo "run        - Run database container for local development."
	@echo "stop       - Stop and remove database container (data is not removed)."

build:
	docker build -t fruitbar-postgres ./

run:
	docker run -d --name fruitbar-postgres-container -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=fruitbar -e POSTGRES_DB=fruitbar -p 5432:5432 fruitbar-postgres

stop:
	-docker stop fruitbar-postgres-container
	-docker rm fruitbar-postgres-container


