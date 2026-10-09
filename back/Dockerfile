FROM golang:1.27-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/server /server

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]