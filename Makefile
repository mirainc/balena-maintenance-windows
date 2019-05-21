all: build run

build:
	@docker build -t mirainc/balena-maintenance-windows .

run:
	@docker run mirainc/balena-maintenance-windows

run-local:
	@dep ensure
	@go build
	@./balena-maintenance-windows

push-docker-hub:
	@docker push mirainc/balena-maintenance-windows:$(tag)
