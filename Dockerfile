FROM golang:1.17-alpine as build

WORKDIR /build

COPY go.mod go.sum ./

COPY pkg pkg

RUN go mod download

COPY . .

ENV CGO_ENABLED=0

RUN go build -ldflags "-s -w" -o _output/domain_exporter ./cmd/main.go

FROM alpine:3.15

COPY --from=build /build/_output/ /usr/local/bin/

ENTRYPOINT [ "domain_exporter" ]
