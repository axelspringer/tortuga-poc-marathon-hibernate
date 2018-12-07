# BUILD GO
FROM golang:alpine as go_builder
LABEL maintainer="jan.michalowsky@axelspringer.com"

# Install git + SSL ca certificates
RUN apk update && apk add git && apk add ca-certificates && apk add make

COPY . $GOPATH/src/github.com/axelspringer/tortuga-poc-marathon-hibernate/
WORKDIR $GOPATH/src/github.com/axelspringer/tortuga-poc-marathon-hibernate/

RUN echo $GOPATH
RUN go get -d -v ./...
RUN make build/hiberthon/static
RUN make build/trigger/static

# BUILD
FROM scratch
LABEL maintainer="jan.michalowsky@axelspringer.com"

COPY --from=go_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go_builder /go/src/github.com/axelspringer/tortuga-poc-marathon-hibernate/bin/hiberthon /hiberthon
COPY --from=go_builder /go/src/github.com/axelspringer/tortuga-poc-marathon-hibernate/bin/hiberthon-trigger /hiberthon-trigger

CMD [ "/hiberthon" ]