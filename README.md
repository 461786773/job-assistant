# 求职助手 / Job Assistant

面向 **3–10 年** ToB / 安全 / 合规迁移求职者的过关助手：简历优化 → 面试模拟 → 薪资确认。

**English:** Job Assistant helps mid-level professionals (3–10 years) moving into ToB / security / compliance roles clear three gates—resume optimization, interview simulation, and compensation confirmation—via a Vue + Go web app powered by an OpenAI-compatible LLM.

产品方案见 [`docs/产品方案_v1.md`](docs/产品方案_v1.md)。

## 功能

- 账号体系：注册 / 登录，任务按用户隔离
- Web 工作台（Vue 3）：简历优化 / 面试模拟 / 薪资确认
- Go API + SQLite 持久化（`data/tasks.db`，纯 Go 驱动）；简历 md/txt/docx/PDF（文本型）
- LLM：OpenAI 兼容中转或 DeepSeek 官方

## 本地启动

依赖：Node 18+、Go 1.22+（构建会按需拉 Go toolchain；若本机无 Go，可用 `~/.local/go`）。

```bash
chmod +x scripts/*.sh scripts/deploy/*.sh
./scripts/dev-watch.sh
```

浏览器打开 http://127.0.0.1:5173 ，先注册账号再使用。

一体启动：`./scripts/dev.sh` → http://127.0.0.1:8080

配置：复制 `.env.example` 为 `.env`，填写 `JA_LLM_*`。可选 `JA_JWT_SECRET`（不填则自动生成到 `data/.jwt_secret`）。

## 腾讯云部署打包

```bash
./scripts/package-release.sh
# 产物: dist-release/job-assistant-linux-amd64-*.tar.gz
```

上传解压后：`cp .env.example .env` → 填 Key → `./start.sh`  
详见 [`docs/腾讯云部署.md`](docs/腾讯云部署.md)。

## 目录

```
apps/api   Go 后端
apps/web   Vue 前端
data/      SQLite（tasks.db）与上传文件（gitignore）
docs/      产品方案 / 部署说明
scripts/   开发与打包脚本
```
