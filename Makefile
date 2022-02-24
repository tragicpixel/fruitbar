#################
# DOCUMENTATION #
#################

# TODO: This doesn't quite give me what I want...need to download all the pages, or run godoc server and then wget everything from this page
# worst case, just run regular godoc and display the page
# TODO: check that godoc is installed, if not download and install
# go get golang.org/x/tools/cmd/godoc
docs-godoc:
	godoc -url "http://localhost:6060/pkg/github.com/tragicpixel/fruitbar/" > godoc.html

# TODO: check that swagger is installed, if not download and install
docs-openapi:
	swagger generate spec -o ./swagger-orders.json --tags=orders
	swagger generate spec -o ./swagger-users.json --tags=users

docs: docs-godoc docs-openapi

########
# TEST #
########

# TODO: check that eslint and react plugin is installed, if not download and install
#npm install -g eslint
#npm install -g eslint-plugin-react
test-lintes:
	eslint fruitbar-ui/src/**/*.js fruitbar-ui/src/**/*.jsx

# TODO: check that golangci-lint is installed, if not download and install
test-lintgo:
	golangci-lint run	

test-runtests:
	go test ./...

test-coverage:
	go test -cover ./...

test: test-lintgo test-coverage test

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