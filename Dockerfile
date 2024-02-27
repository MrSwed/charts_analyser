FROM golang:1.21-alpine as builder
WORKDIR /source
COPY cmd cmd
COPY internal internal
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/migrate -o /migrate_app
RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/app -o /server_app
RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/simulator -o /simulator_app

FROM scratch as migrate
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder migrate_app /migrate_app
ENTRYPOINT ["/migrate_app"]

FROM scratch as server_app
EXPOSE 3000
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder server_app /server_app
ENTRYPOINT ["/server_app"]


FROM scratch as simulator
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder simulator_app /simulator_app
ENTRYPOINT ["/simulator_app"]