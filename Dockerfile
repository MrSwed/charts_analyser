FROM golang:1.21-alpine as builder
WORKDIR /source
COPY cmd cmd
COPY internal internal
COPY docs docs
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/migrate -o /migrate_app
RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/app -o /server_app
RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/simulator -o /simulator_app
RUN CGO_ENABLED=0 GOOS=linux go build -o /swagger_app

FROM scratch as migrate
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder migrate_app /migrate_app
ENTRYPOINT ["/migrate_app"]

FROM scratch as server
EXPOSE ${LISTEN_PORT}
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder server_app /server_app
ENTRYPOINT ["/server_app"]


FROM scratch as simulator
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder simulator_app /simulator_app
ENTRYPOINT ["/simulator_app"]

FROM scratch as swagger
EXPOSE ${SWAGGER_PORT}
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder swagger_app /swagger_app
ENTRYPOINT ["/swagger_app"]
