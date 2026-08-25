# 运维与部署指南（本机或云服务器）

更完整的恢复、备份和账号轮换步骤见 `docs/ops_runbook.md`。

## 依赖与端口

- Docker / Docker Compose
- Web：宿主机 `80`，提供前端和 `/api/` 反向代理
- MySQL：宿主机 `127.0.0.1:13306`，不对公网开放

当前部署按项目约束继续使用 HTTP：

- 本机：`http://127.0.0.1/`
- 公网：`http://<your-server-ip-or-domain>/`

## 首次配置

仓库不提供默认登录账号、明文密码或 JWT 密钥。首次启动前创建本地 `.env`：

```bash
cp .env.example .env
go run ./tools/hash_password
openssl rand -hex 32
```

运行密码哈希工具后，在标准输入中输入新密码并回车。将输出的 bcrypt 哈希和随机 JWT 密钥分别填入 `.env`：

```dotenv
CMS_AUTH_USERNAME=your-admin-name
CMS_AUTH_PASSWORD_HASH='$2a$10$replace-with-generated-hash'
CMS_AUTH_JWT_SECRET=replace-with-generated-random-secret
```

`CMS_AUTH_PASSWORD_HASH` 应保留单引号，避免 bcrypt 哈希中的 `$` 被 Compose 解析。`.env` 已被 Git 忽略，不要提交真实凭据。

启动前可先校验 Compose 配置：

```bash
docker compose config --quiet
```

## 启动与验证

```bash
docker compose up -d --build
docker compose ps
curl -i http://127.0.0.1/api/health
```

健康检查会实际探测数据库连接。后端首次启动会执行兼容迁移和历史快照回填，不会清理积分明细。

## 从旧版本升级到 2.0.0

先按 `docs/ops_runbook.md` 备份数据库，并确保 MySQL 容器正在运行。随后在仓库根目录执行版本迁移脚本：

```bash
docker compose exec -T mysql \
  sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE"' \
  < deploy/sql/dev-2.0.0.sql
```

脚本可以重复执行，不会清理积分明细。它会增加软删除、历史快照和计分幂等结构，并在建立公平点名唯一索引前去除重复的临时点名记录。迁移完成后再执行 `docker compose up -d --build` 更新服务。

登录示例中的值应替换为 `.env` 中设置的账号和实际密码：

```bash
curl -sS -H 'Content-Type: application/json' \
  -d '{"username":"your-admin-name","password":"your-password"}' \
  http://127.0.0.1/api/auth/login
```

## 凭据轮换

修改用户名或密码时，重新生成 bcrypt 哈希并更新 `.env`，然后重建后端容器：

```bash
docker compose up -d --force-recreate cms-api
```

修改 `CMS_AUTH_JWT_SECRET` 会立即使所有旧 token 失效，之后需要重新登录。这是密钥轮换的预期行为。

## 数据保留与备份

- 学生删除为软删除，历史积分仍保留。
- 积分明细永久保留，只有主动撤销才删除。
- MySQL 数据位于具名卷 `mysql_data`，普通停止、重建容器或机器重启不会删除数据。
- 升级和凭据轮换前建议按 `docs/ops_runbook.md` 先做逻辑备份。

## 日志与停止

```bash
docker compose logs -f cms-api
docker compose logs -f mysql
docker compose down
```

`docker compose down -v` 会永久删除 MySQL 数据卷；除非明确要清空所有数据，否则不要执行。
