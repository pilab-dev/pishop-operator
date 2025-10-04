FROM golang:1.25.1-alpine AS builder

# Build arguments for version information
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_DATE

LABEL org.opencontainers.image.source=https://github.com/pilab-dev/pishop-operator
LABEL org.opencontainers.image.description="PiShop Operator for PR Environment Management"
LABEL org.opencontainers.image.licenses="Progressive Innovation LAB. (c) 2025"
LABEL org.opencontainers.image.authors="Paál Gyula <gyula@pilab.hu>"
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.vendor="Progressive Innovation LAB."


# Set build-time labels
LABEL version=${VERSION}
LABEL commit=${COMMIT_SHA}
LABEL build-date=${BUILD_DATE}

RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT_SHA}' -X 'main.BuildDate=${BUILD_DATE}'" \
    -o pishop-operator ./operator

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/pishop-operator .
ENTRYPOINT ["./pishop-operator"]
CMD ["--help"]
