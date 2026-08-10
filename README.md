# 求职助手 / Job Assistant

面向 **3–10 年** ToB / 安全 / 合规人群的**职场心理教练**。主路径：链接登录 → 心理评测与 AI 分析 → 确认诉求 → 回看评估 → AI 疏导（可预约私人辅导）→ 按需求职三关或晋升/沟通服务 → 问卷跟踪。

**English:** Job Assistant is a workplace psychological coach: link login, first assessment with AI analysis, need confirmation, reviewable results, AI counseling (plus private counseling booking), then optional job-search gates or promotion/communication services, with questionnaire-based wellbeing tracking.

产品方案见 [`docs/产品方案_v1.md`](docs/产品方案_v1.md)（主故事线 §0.4）；问卷见 [`docs/初次心理评估问卷.md`](docs/初次心理评估问卷.md)、[`docs/三分钟自评表.md`](docs/三分钟自评表.md)。

## 功能

- **主故事线**：链接登录 · 初次评测 + AI 分析 · 诉求确认 · 我的评估 · AI 疏导 · 私人辅导预约 · 问卷跟踪
- **求职子流程**：人事 → 业务 → 谈薪
- **晋升 / 沟通**：对应场景服务
- 账号受控访问；Vue 3 + Go + SQLite；LLM 服务端配置

## 本地启动

依赖：Node 18+、Go 1.22+（构建会按需拉 Go toolchain；若本机无 Go，可用 `~/.local/go`）。

```bash
chmod +x scripts/*.sh scripts/deploy/*.sh
./scripts/dev-watch.sh
```

浏览器打开 http://127.0.0.1:5173 ，输入用户名后走主故事线（新用户先评测与诉求确认）。

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
