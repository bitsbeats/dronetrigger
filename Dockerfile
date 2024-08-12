FROM golang:1-alpine as builder

RUN true \
    && apk add --no-cache ca-certificates

ADD . /app
WORKDIR /app

ENV CGO_ENABLED=0

RUN go build -o /app/app ./cmd/dronetrigger-web

# ---

FROM scratch
COPY --from=builder /app/app /app
COPY --from=builder /etc/ssl/cert.pem /etc/ssl/cert.pem
ENTRYPOINT ["/app"]
