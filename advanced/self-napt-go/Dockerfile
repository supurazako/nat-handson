FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY prototype ./prototype
RUN go build -o /out/naptd ./prototype/cmd/naptd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates iproute2 iptables conntrack-tools
COPY --from=build /out/naptd /usr/local/bin/naptd
ENTRYPOINT ["/usr/local/bin/naptd"]
