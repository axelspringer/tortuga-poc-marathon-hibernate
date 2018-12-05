# BUILD GO
FROM golang:alpine as go_builder
LABEL maintainer="jan.michalowsky@axelspringer.com"

# Install git + SSL ca certificates
RUN apk update && apk add git && apk add ca-certificates && apk add make gcc musl-dev

COPY . $GOPATH/src/github.com/axelspringer/tortuga-poc-marathon-hibernate/
WORKDIR $GOPATH/src/github.com/axelspringer/tortuga-poc-marathon-hibernate/

RUN echo $GOPATH
RUN go get -d -v ./...
RUN make build/hiberthon
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-w -s" -o hibernate main.go

# BUILD
FROM scratch
LABEL maintainer="jan.michalowsky@axelspringer.com"

COPY --from=go_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go_builder /go/src/github.com/axelspringer/tortuga-poc-marathon-hibernate/bin/hiberthon /hiberthon

ENTRYPOINT ["/hiberthon"]