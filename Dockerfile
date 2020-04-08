FROM alpine:3.11 as certs

RUN apk add --no-cache ca-certificates

# ---

FROM golang:1.14 as builder

WORKDIR /src
ADD . .
RUN true \
	&& GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor ./cmd/dronetrigger-web \
	&& strip dronetrigger-web

# ---

FROM busybox

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /src/dronetrigger-web /bin/dronetrigger-web

CMD dronetrigger-web
