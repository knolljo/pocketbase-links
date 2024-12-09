# Distroless setup
FROM golang:1.23.3 as build

WORKDIR /go/src/app
COPY go.mod go.sum main.go /go/src/app/
COPY migrations/ /go/src/app/migrations

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/app

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian12
COPY --from=build /go/bin/app /
CMD ["/app", "serve", "--http", "0.0.0.0:8090"]

EXPOSE 8090
# syntax=docker/dockerfile:1
