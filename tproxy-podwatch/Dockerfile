FROM alpine:3.6
RUN apk add --update curl iptables
COPY gopath/bin/tproxy-podwatch /tproxy-podwatch
ENTRYPOINT ["/tproxy-podwatch"]