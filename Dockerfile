FROM golang:1.26 as Builder

WORKDIR /pprof-web
COPY . .
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" && chmod +x pprof-web

FROM alpine:3.23

RUN apk add --no-cache graphviz
COPY --from=Builder /pprof-web/pprof-web /

CMD ["/pprof-web"]
