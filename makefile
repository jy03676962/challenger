all: build run

build:
	./build.sh
run:
	cd ./server && go run main.go
