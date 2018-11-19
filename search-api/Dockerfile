FROM golang:1.10.0

RUN adduser --disabled-password --gecos '' api
USER api

RUN go get github.com/gin-gonic/gin
RUN go get github.com/olivere/elastic
RUN go get github.com/teris-io/shortid 

WORKDIR /go/src/app
COPY . .

RUN go install -v ./...

CMD [ "app" ]
