FROM golang:latest AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=build /go/bin/app /app/app
CMD ["/app/app"]