FROM golang:latest as Builder

WORKDIR /pprof-web
COPY . .
ENV CGO_ENABLED=0
RUN go build && chmod +x pprof-web

FROM alpine:3.23

RUN apk add --no-cache graphviz
COPY --from=Builder /pprof-web/pprof-web /

CMD ["/pprof-web"]
