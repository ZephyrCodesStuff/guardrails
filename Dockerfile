FROM golang:latest AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-s -w' -o /go/bin/app

FROM scratch

WORKDIR /app
# nobody:nobody
USER 65534:65534

COPY --from=build /go/bin/app /app/app
CMD ["/app/app"]