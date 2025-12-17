FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY portainer-cli /usr/local/bin/portainer-cli

ENTRYPOINT ["/usr/local/bin/portainer-cli"]
