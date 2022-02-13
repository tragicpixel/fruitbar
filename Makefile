#IMAGE_NAME := orders
SHELL=bash
#shell=cmd (perhaps change things to use cmd, or also provide a bash version??)

# Modufy
ifdef OS
   RM = del /Q
   FixPath = $(subst /,\,$1)
else
   ifeq ($(shell uname), Linux)
      RM = rm -f
      FixPath = $1
   endif
endif

# $(RM) $(call FixPath,objs/*)

BUILDDIR = cmd

ORDERSDIR = orders
USERSDIR = users

BUILDDIRS = $(ORDERSDIR) $(USERSDIR)

DBDIR = database

SUBDIRS = $(ORDERSDIR) $(USERSDIR) $(DBDIR)

build-dev:
	docker-compose build

run:
	docker-compose up -d

stop:
	docker-compose kill
	docker-compose stop
	docker-compose down --rmi local
	docker-compose rm -f

api-docs:
	swagger generate spec -o ./swagger-orders.json --tags=orders
	swagger generate spec -o ./swagger-users.json --tags=users

lint-go:
	golangci-lint run
# need to modify this to use config to ignore 'structcheck' 'body' not used errors

# this works: just use this as an example (need to be in bash shell, wont work for windows--separate windows version??)
all:
	cd $(BUILDDIR)
	for i in $(BUILDDIRS);\
	do \
		( cd $$i ; make build); \
	done