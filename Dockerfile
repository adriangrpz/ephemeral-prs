# Stage 1: Build static Go binary
FROM golang:1.26-bookworm AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server .

# Stage 2: Minimal non-root runtime image
FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /
COPY --from=builder /out/server /server

ENV PORT=8080
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/server"]