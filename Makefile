
#
# Makefile - for helping build and running from gen-blog.go
# (C) Philip Schlump, 2013.
#

all: ./bin/gen-makefile ./bin/gen-blog ./bin/bf ./bin/inc-er quiet ./bin/go-server ./bin/go-serv3 ./bin/go-lfp w-watch ./bin/go-ser-server ./bin/go-ser-edit

install:
	cp ./bin/* ~/bin

run_server: ./bin/go-server
	./bin/go-server 8765 ./_site &
	gen-blog

run_8764_server:
	go run go-ser-edit.go &

w-watch: w-watch.go
	go build w-watch.go

./bin/go-ser-server: go-ser-server.go
	go build go-ser-server.go
	@mv go-ser-server bin

./bin/go-ser-edit: go-ser-edit.go
	go build go-ser-edit.go
	@mv go-ser-edit bin

./bin/go-lfp: go-lfp.go
	go build go-lfp.go
	@mv go-lfp bin

./bin/go-serv3: go-serv3.go
	go build go-serv3.go
	@mv go-serv3 bin

./bin/gen-makefile: gen-makefile.go
	go build gen-makefile.go
	@mv gen-makefile bin

./bin/go-server: go-server.go
	go build go-server.go
	@mv go-server bin

./bin/bf: bf.go
	go build bf.go
	@mv bf bin

./bin/inc-er: inc-er.go
	go build inc-er.go
	@mv inc-er bin

./bin/gen-blog: gen-blog.go
	go build gen-blog.go
	@mv gen-blog bin


clean:
	rm -f ./tmp/* ./posts/*
	echo "clean"


# Copy in static resources
finialize:
	( cd ./static ; rsync -va . ../_site )

# Publish the site to the local world
push_up:
	( cd _site ; rsync -va --delete . /usr/local/nginx/html/t1 ) >.rsync-push-up.log

# Publish the site to the local world
publish: push_up
	gen-blog -g
	git commit -m "Published to world on $(date)" .
	git push origin master
	gen-blog 

push_up_old:
	( cd _site ; tar  -cf - . ) | ( cd /usr/local/nginx/html/t1 ; tar -xvf - )
	echo git push origin master

# Setup this directory for running stuff
setup:
	-mkdir -p ./tmp ./_site ./res ./_site/posts ./_site/data ./_site/img ./_site/css ./_site/js ./draft ./bak ./log
	go get code.google.com/p/gorilla/mux
	go get github.com/russross/blackfriday

quiet:
	@echo "quite" >/dev/null

