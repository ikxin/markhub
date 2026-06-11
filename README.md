# Markhub

Markhub is a small Gin service that proxies avatar and favicon images. Favicon
images are normalized with libvips through govips.

## Requirements

- Go 1.25 or newer
- libvips and pkg-config

On macOS:

```bash
brew install vips pkg-config
```

On Debian/Ubuntu:

```bash
apt-get update
apt-get install -y libvips-dev pkg-config
```

## Development

Run the service on `http://localhost:3000`:

```bash
go run ./cmd/markhub
```

Use `HOST` and `PORT` to override the bind address:

```bash
HOST=127.0.0.1 PORT=8080 go run ./cmd/markhub
```

Run tests:

```bash
go test ./...
```

## Routes

- `GET /github/:id?id`
- `GET /github/:username`
- `GET /gravatar/:hashOrEmail`
- `GET /qq/:number`
- `GET /telegram/:user`
- `GET /opencollective/:user`
- `GET /favicon/:host`

Responses use `Content-Type: image/webp` and
`Cache-Control: max-age=2592000`. Images are normalized to 100x100 WebP by
default after fetching a high-resolution source image from each provider. Use
the `size` or `s` query parameter to change the WebP output size up to
2048x2048. Failed upstream requests return the matching fallback image as WebP.

## Docker

Build and run:

```bash
docker build -t markhub .
docker run --rm -p 3000:3000 markhub
```
