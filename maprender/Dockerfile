FROM golang:1.20 AS builder
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY maprender/ /go/src/app/maprender/
WORKDIR /go/src/app/maprender/
RUN CGO_ENABLED=0 go build

FROM gcr.io/distroless/static-debian11 AS runtime
COPY --from=builder /go/src/app/maprender/maprender /usr/local/bin/maprender
ENTRYPOINT ["/usr/local/bin/maprender"]
