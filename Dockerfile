FROM balenalib/intel-nuc-golang:1.12.4-stretch-20190511

RUN apt-get update
RUN apt-get install curl

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/app

COPY . .
RUN dep ensure
RUN go install -v ./...

CMD ["app"]
