all: build run

build:
	@docker build -t balena-maintenance-windows .

run:
	@docker run balena-maintenance-windows

run-local:
	@dep ensure
	@go build
	@./balena-maintenance-windows
