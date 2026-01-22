# Build the binary on the golang image
FROM golang:1.25 AS build
WORKDIR /go/src/app

# Dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-extldflags=-static -s" -buildvcs=false -o /go/bin/service ./cmd/server/main.go

# Run the binary on distroless non-root
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /go/bin/service /app/service
EXPOSE 8080

ENTRYPOINT ["/app/service"]
