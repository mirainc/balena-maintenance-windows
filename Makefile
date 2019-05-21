all: build run

build:
	@docker build -t mirainc/balena-maintenance-windows .

run:
	@touch .env
	@docker run --env-file .env --name balena-maintenance-windows mirainc/balena-maintenance-windows

run-local:
	@dep ensure
	@go build
	@./balena-maintenance-windows

push-docker-hub: build
	@docker push mirainc/balena-maintenance-windows:$(tag)
