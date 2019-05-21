FROM balenalib/intel-nuc-golang:1.12.4-stretch-20190511 as build

RUN apt-get update
RUN apt-get install curl

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/app

COPY ./Gopkg.lock .
COPY ./Gopkg.toml .

RUN dep ensure

COPY ./main.go .
RUN go install -v ./...

FROM balenalib/intel-nuc-debian:stretch-run-20190511
WORKDIR /go/src/app

COPY --from=build /go/bin/app /usr/local/bin/app

CMD ["app"]
