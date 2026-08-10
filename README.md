# 求职助手 / Job Assistant

面向 **3–10 年** ToB / 安全 / 合规迁移求职者的**职场心理教练**：稳住状态 → 理清选择 → 过关训练（人事 / 业务 / 谈薪）。

**English:** Job Assistant is a workplace psychology coach for mid-level professionals (3–10 years) in ToB / security / compliance moves—stabilize under pressure, clarify choices, then train the three hiring gates when needed.

产品方案见 [`docs/产品方案_v1.md`](docs/产品方案_v1.md)。

## 功能

- **职场心理教练**：求职 / 晋升 / 沟通三场景会话（含行动卡与危机转介）
- **心理健康跟踪**：压力/情绪打卡 + 近 7/30 日摘要（仅本人可见）
- **过关训练**：人事关评分改写 · 业务关模拟 · 谈薪关对照
- 账号：用户名隔离（初期无密码）；Vue 3 + Go + SQLite；LLM 服务端配置

## 本地启动

依赖：Node 18+、Go 1.22+（构建会按需拉 Go toolchain；若本机无 Go，可用 `~/.local/go`）。

```bash
chmod +x scripts/*.sh scripts/deploy/*.sh
./scripts/dev-watch.sh
```

浏览器打开 http://127.0.0.1:5173 ，输入用户名后进入**教练工作台**。

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
apps/api   Go 后端（coach / wellbeing / tasks 三关）
apps/web   Vue 前端
data/      SQLite（tasks.db）与上传文件（gitignore）
docs/      产品方案 / 问卷 / 部署说明
scripts/   开发与打包脚本
```
