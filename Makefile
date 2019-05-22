all: build run

build: clean
	@docker build -t mirainc/balena-maintenance-windows .

build-test:
	@docker build -t mirainc/balena-maintenance-windows-test -f Dockerfile.test .

clean:
	-@docker kill balena-maintenance-windows
	-@docker rm balena-maintenance-windows

run:
	@touch .env
	@docker run -t -i --env-file .env --name balena-maintenance-windows mirainc/balena-maintenance-windows

run-local: build-local
	@./balena-maintenance-windows

build-local:
	@dep ensure
	@go build

push-docker-hub: build
	@docker push mirainc/balena-maintenance-windows:$(tag)

test-local: build-local
	@go test ./...

test: build-test
	@docker run -t -i --name balena-maintenance-windows-test mirainc/balena-maintenance-windows-test
