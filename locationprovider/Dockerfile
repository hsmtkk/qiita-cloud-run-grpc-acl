FROM golang:1.20 AS builder
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY locationprovider/ /go/src/app/locationprovider/
WORKDIR /go/src/app/locationprovider/
RUN CGO_ENABLED=0 go build

FROM gcr.io/distroless/static-debian11 AS runtime
COPY --from=builder /go/src/app/locationprovider/locationprovider /usr/local/bin/locationprovider
ENTRYPOINT ["/usr/local/bin/locationprovider"]
