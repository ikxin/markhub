FROM golang:1.26-bookworm AS builder

WORKDIR /app

RUN apt-get update \
  && apt-get install -y --no-install-recommends libvips-dev pkg-config \
  && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /out/markhub ./cmd/markhub

FROM debian:bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates libvips42 \
  && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/markhub /usr/local/bin/markhub

EXPOSE 3000

ENV HOST=0.0.0.0
ENV PORT=3000

CMD ["markhub"]
