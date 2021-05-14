FROM golang:alpine AS build

WORKDIR /src/
COPY main.go go.* /src/

RUN go mod download

RUN CGO_ENABLED=0 go build -o /bin/immotrakt

FROM scratch
COPY --from=build /bin/immotrakt /bin/immotrakt
COPY config.yml ./
# NB: this pulls directly from the upstream image, which already has ca-certificates:
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/bin/immotrakt"]