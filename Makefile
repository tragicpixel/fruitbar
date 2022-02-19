#################
# DOCUMENTATION #
#################

# TODO: generate and store go docs???

# TODO: check that swagger is installed, if not download and install
docs-openapi:
	swagger generate spec -o ./swagger-orders.json --tags=orders
	swagger generate spec -o ./swagger-users.json --tags=users

docs: docs-openapi

########
# TEST #
########

# TODO: Lint for javascript? (or RSX)

test-runtests:
	go test ./...

test-coverage:
	go test -cover ./...

# TODO: check that golangci-lint is installed, if not download and install
test-lintgo:
	golangci-lint run

test: test-coverage test test-lintgo

#########
# BUILD #
#########

# TODO: multi-stage builds
# TODO: version number tags for builds

build-verifydeps:
	go mod tidy
	go mod verify

build-vendor:
	go mod vendor

build-images:
	docker-compose build

build: build-verifydeps build-vendor build-images

#######
# RUN #
#######

# TODO: run/stop by multi-stage
# ->> perhaps make a generic run command with variables for stage, and then indiv command utilizing the generic command

run:
	docker-compose up -d

stop:
	docker-compose kill
	docker-compose stop
	docker-compose down --rmi local
	docker-compose rm -f