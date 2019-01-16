FROM golang:1.11-alpine3.8 as builder

##
# User builder image to minimize final image.
##
WORKDIR /github.com/dgra/owlet-golang
COPY . /github.com/dgra/owlet-golang

# Install dependencies, and build our executable.
RUN apk update && \
  apk add --no-cache \
    git  \
    wget && \
  rm -rf /var/cache/apk/* && \
  go mod vendor && \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /main

##
# Using scratch as a base means I need to compile the go code as a single statically linked executable.
##
FROM scratch

# Only three files are needed.
# ca-certificats for ssl connections, our config and our executable.
COPY --from=builder /github.com/dgra/owlet-golang/config.json /config.json
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /main /main

CMD ["/main"]
