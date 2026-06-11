# Markhub

Markhub 是一个基于 Gin 的小型图片代理服务，用于代理头像和网站图标。网站图标会通过 govips 调用 libvips 进行标准化处理。

## 环境要求

- Go 1.25 或更高版本
- libvips 和 pkg-config

macOS:

```bash
brew install vips pkg-config
```

Debian/Ubuntu:

```bash
apt-get update
apt-get install -y libvips-dev pkg-config
```

## 开发

启动服务，默认地址为 `http://localhost:3000`:

```bash
go run ./cmd/markhub
```

可以通过 `HOST` 和 `PORT` 覆盖监听地址:

```bash
HOST=127.0.0.1 PORT=8080 go run ./cmd/markhub
```

运行测试:

```bash
go test ./...
```

## 路由

- `GET /github/:id?id`
- `GET /github/:username`
- `GET /gravatar/:hashOrEmail`
- `GET /qq/:number`
- `GET /telegram/:user`
- `GET /opencollective/:user`
- `GET /favicon/:host`

响应会设置 `Cache-Control: max-age=2592000`。服务会先从各 provider 获取高分辨率源图，再默认标准化为 100x100 WebP 图片。

可以通过受支持的后缀指定输出格式: `.webp`、`.jpg`、`.jpeg`、`.png`、`.avif`、`.gif`。无后缀时，可以通过 `format` 或 `fmt` 查询参数指定输出格式。优先级为: 路径后缀、`format`、`fmt`、默认 WebP。非法格式值会继续尝试下一优先级，全部非法时回退为 WebP。

可以通过 `size` 或 `s` 查询参数修改输出尺寸，最大为 2048x2048。上游请求失败时，会返回对应 provider 的 fallback 图片，并使用同样的输出格式规则。

## Docker

构建并运行:

```bash
docker build -t markhub .
docker run --rm -p 3000:3000 markhub
```
