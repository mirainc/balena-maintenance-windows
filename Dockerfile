FROM balenalib/intel-nuc-golang:1.12.4-stretch-20190511 as build

RUN apt-get update
RUN apt-get install curl

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/balena-maintenance-windows

COPY ./Gopkg.lock .
COPY ./Gopkg.toml .
COPY ./main.go .

RUN dep ensure
RUN go install -v ./...

FROM balenalib/intel-nuc-debian:stretch-run-20190511
WORKDIR /go/src/balena-maintenance-windows

COPY --from=build /go/bin/balena-maintenance-windows /usr/local/bin/balena-maintenance-windows
COPY ./start-debian.sh .
RUN chmod +x ./start-debian.sh

ENV TIMEZONE UTC

CMD ["./start-debian.sh"]
