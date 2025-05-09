# Docker 镜像发布到 GitHub Container Registry (ghcr.io)

| Feature     | Status | Date       |
| ----------- | ------ | ---------- |
| Tag 触发构建并推送 | Done   | 2026-03-02 |

---

## Table of Contents

- [概述](#概述)
- [触发条件](#触发条件)
- [Workflow 与镜像命名](#workflow-与镜像命名)
- [如何发布](#如何发布)
- [如何拉取与运行](#如何拉取与运行)
- [说明](#说明)

---

## 概述

通过 **GitHub Actions** 在推送符合规则的 **Git tag** 时，自动构建当前仓库的 Docker 镜像（多阶段 [Dockerfile](../../Dockerfile)）并推送到 **GitHub Container Registry (ghcr.io)**。最终镜像内仅包含二进制与 `migrations/`，不包含源码。无需在仓库中配置额外 Secret，使用 `GITHUB_TOKEN` 即可推送。

---

## 触发条件

- **事件**：`push` 到 **tags**
- **Tag 格式**：`v*`（例如 `v1.0.0`、`v0.1.0`）
- 推送此类 tag 后，Actions 会自动运行构建并推送镜像。

---

## Workflow 与镜像命名

- **Workflow 文件**：[.github/workflows/publish-docker.yml](../../.github/workflows/publish-docker.yml)
- **镜像命名**：`ghcr.io/<owner>/uim-go:<tag>`
  - `<owner>` 为仓库所属组织或用户名（与 `github.repository` 一致，如 `convexwf/uim-go` 则镜像名为 `ghcr.io/convexwf/uim-go`）
  - `<tag>` 与 Git tag 一致（如推送 `v1.0.0` 则镜像 tag 为 `v1.0.0`）

---

## 如何发布

1. 在本地为要发布的提交打 tag（语义化版本）：
   ```bash
   git tag v1.0.0
   ```
2. 推送 tag 到 GitHub：
   ```bash
   git push origin v1.0.0
   ```
3. 在仓库 **Actions** 页查看 workflow 运行状态；成功完成后，镜像会出现在仓库 **Packages** 中。

---

## 如何拉取与运行

- **拉取镜像**（将 `<owner>` 替换为实际组织/用户名）：
  ```bash
  docker pull ghcr.io/<owner>/uim-go:v1.0.0
  ```
- 若镜像为**私有**，需先登录：
  ```bash
  echo "YOUR_GITHUB_PAT" | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin
  ```
- **运行示例**（端口、环境变量等按需调整）：
  ```bash
  docker run -d --name uim-server -p 8080:8080 -p 8081:8081 \
    -e POSTGRES_HOST=... -e POSTGRES_PORT=5432 ... \
    ghcr.io/<owner>/uim-go:v1.0.0
  ```

---

## 说明

- 镜像由多阶段 Dockerfile 构建，**最终层不包含 Go 源码**，仅包含编译后的 `uim-server` 二进制与 `migrations/` 目录。
- 推送至 ghcr.io 使用仓库默认的 `GITHUB_TOKEN`，无需在 Settings 中配置额外 Secret。
- 验收要点（由用户自行执行）：推送 tag 后确认 Actions 成功、Packages 中出现镜像、本地 pull 并运行确认行为正常且镜像内无源码。
