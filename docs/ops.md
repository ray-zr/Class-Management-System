# 运维与部署指南（本机云服务器）

更详细的运维手册（含宕机恢复、备份恢复、修改账号密码）见：`docs/ops_runbook.md`。

## 依赖

- Docker / Docker Compose

## 端口说明

默认 Compose 端口映射：

- MySQL：宿主机 `13306` -> 容器 `3306`
- Web（前端 + API 反代）：宿主机 `80` -> 容器 `80`

（为避免与宿主机常见占用端口冲突。）

## 启动

在仓库根目录：

```bash
docker compose up -d
docker compose ps
```

健康检查：

```bash
curl -i http://127.0.0.1/api/health
```

浏览器访问：

- 本机：`http://127.0.0.1/`
- 公网：`http://<your-server-ip-or-domain>/`

前端功能入口：

- 学生名单：支持 Excel 导入（`.xlsx`）
- 积分录入：最近使用置顶、加/减按钮色彩语义（绿/红）
- 随机点名：支持公平模式；点名结果可一键跳转到积分录入并自动选中该学生
- 排行榜：支持筛选与导出 Excel（走鉴权下载）

## 登录账号

默认账号来自配置文件 `backend/etc/cms-api.docker.yaml`：

- 用户名：`teacher`
- 密码：`teacher`

登录：

```bash
curl -sS -H 'Content-Type: application/json' \
  -d '{"username":"teacher","password":"teacher"}' \
  http://127.0.0.1/api/auth/login
```

## 配置

API 容器内读取：`/app/etc/cms-api.yaml`。

Compose 通过只读 volume 挂载：

- `./backend/etc/cms-api.docker.yaml:/app/etc/cms-api.yaml:ro`

排行榜高亮阈值默认值来自配置：

- `App.RankingTopN`：未指定 `topN` 参数时使用（默认 5）

## 日志

```bash
docker compose logs -f cms-api
docker compose logs -f mysql
```

## 停止与清理

```bash
docker compose down
```

清理 MySQL 数据（删除数据卷，数据不可恢复）：

```bash
docker compose down -v
```
