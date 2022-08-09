FROM golang:1.17.2-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /src/bin/grpcwebapp -tags=jsoniter,nomsgpack .

FROM scratch AS bin
WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /src/bin/grpcwebapp .

CMD ["/app/grpcwebapp"]
