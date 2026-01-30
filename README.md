# OpenClaw Setup

一个轻量的 Go + React 工具，用于生成 OpenClaw 的 `openclaw.json` 与 `.env`，并可选择性重启容器。

当前版本：`0.0.1`

## 本地开发

前端：

```bash
cd web
npm install
npm run dev
```

后端：

```bash
go run ./cmd/server
```

默认访问：`http://127.0.0.1:5173/setup`

## CLI 初始化

在 compose 目录执行（读取当前目录下 `.env` 中的 PROVIDER / API_KEY / MODEL）：

```bash
./openclaw-setup init
```

生成文件：
- `data/conf/openclaw.json`
- `data/conf/.env`

## 构建

```bash
make build
```

Linux amd64：

```bash
make build-linux
```

Linux arm64：

```bash
make build-linux-arm64
```
