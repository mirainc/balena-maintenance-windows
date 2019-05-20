all: build run

build:
	@docker build -t balena-maintenance-windows .

run:
	@docker run balena-maintenance-windows
