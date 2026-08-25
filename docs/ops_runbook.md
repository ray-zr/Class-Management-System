# 运维手册（Runbook）

本文档面向“把系统跑在一台云服务器上”的运维场景，覆盖：启动/停止、故障恢复（机器宕机/重启/服务异常）、备份与恢复、账号密码修改、常见排障。

> 部署形态：Docker Compose（MySQL + 后端 API + Nginx 静态前端与 `/api/` 反代）。

---

## 0. 关键事实（先记住）

### 端口

- Web（前端 + API 反代）：宿主机 `80`（对公网开放）
- MySQL：宿主机 `13306`（默认仅绑定 `127.0.0.1`，不建议对公网开放）

### 数据是否会丢

- MySQL 数据存储在 Docker **具名卷**：`mysql_data`。
- 只要你没有删除该卷（例如 `docker compose down -v`），**删容器/重启机器不会丢数据**。

### 登录凭据

- 仓库没有默认账号、明文密码或可用的 JWT 密钥。
- `CMS_AUTH_USERNAME`、`CMS_AUTH_PASSWORD_HASH`、`CMS_AUTH_JWT_SECRET` 必须由部署环境提供，缺失或不安全时后端会拒绝启动。
- 密码只以 bcrypt 哈希参与校验，无法从哈希反查明文。

---

## 1. 首次部署（云服务器）

### 1.1 前置依赖

- Docker
- Docker Compose（`docker compose` 子命令可用）

### 1.2 拉取并启动

在服务器上：

```bash
git clone <your-repo-url>
cd Class-Management-System

cp .env.example .env
go run ./tools/hash_password
openssl rand -hex 32

# 将用户名、bcrypt 输出和随机密钥填入 .env 后再启动
docker compose config --quiet

docker compose up -d --build
docker compose ps
```

### 1.3 开放安全组 / 防火墙

必须放行：

- 入站 TCP `80`

不建议开放（除非你明确知道为什么要开放）：

- 入站 TCP `13306`（数据库）

### 1.4 验证

在服务器本机：

```bash
curl -i http://127.0.0.1/api/health
```

在任意外网机器：

```bash
curl -i http://<公网IP>/api/health
```

浏览器访问：

- `http://<公网IP>/`

---

## 2. 日常运维命令速查

在仓库根目录执行：

```bash
docker compose ps
docker compose logs -f web
docker compose logs -f cms-api
docker compose logs -f mysql
```

停止：

```bash
docker compose down
```

重启全部服务：

```bash
docker compose restart
```

只重启某个服务：

```bash
docker compose restart web
docker compose restart cms-api
docker compose restart mysql
```

更新代码后重新构建/启动（通常用于升级后端镜像）：

```bash
git pull
docker compose up -d --build
```

从旧版本首次升级到 2.0.0 时，应在重新构建服务前执行数据库迁移：

```bash
docker compose exec -T mysql \
  sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE"' \
  < deploy/sql/dev-2.0.0.sql
```

执行前必须先做第 6 节的逻辑备份。`dev-2.0.0.sql` 可重复执行，保留积分历史；其中只会清理同一轮次、同一学生的重复临时点名记录，以便建立唯一索引。

---

## 3. 机器宕机 / 服务器重启后，如何恢复系统

### 3.1 结论

`docker-compose.yml` 已为所有服务设置了 `restart: unless-stopped`。

- 机器重启后：Docker daemon 启动后，会自动拉起这些容器。
- 如果你之前手动 `docker compose down` 过，容器不在了，那么需要你再 `docker compose up -d`。

### 3.2 标准恢复流程（推荐按顺序）

1) 登录服务器，确认 Docker 正常：

```bash
docker info >/dev/null
```

2) 进入仓库目录并确认服务状态：

```bash
cd Class-Management-System
docker compose ps
```

3) 若服务未启动或异常，直接重拉起：

```bash
docker compose up -d
docker compose ps
```

4) 健康检查：

```bash
curl -i http://127.0.0.1/api/health
```

5) 外网验证（从你本地电脑）：

```bash
curl -i http://<公网IP>/api/health
```

### 3.3 如果 MySQL 起不来

1) 看日志：

```bash
docker compose logs --tail=200 mysql
```

2) 常见原因：磁盘满（Docker/卷无法写入）。先清理空间，再重启：

```bash
df -h
docker system df
docker system prune -f
docker compose restart mysql
```

> 注意：`docker system prune` 会删掉未使用镜像/容器/网络，不会删具名卷；但不要在不理解后果的情况下频繁使用。

### 3.4 如果 web(Nginx) 正常但登录/接口失败

1) 看 `cms-api` 日志：

```bash
docker compose logs --tail=200 cms-api
```

2) 看接口健康：

```bash
curl -i http://127.0.0.1/api/health
```

3) 重启后端：

```bash
docker compose restart cms-api
```

---

## 4. 如何修改登录用户名与密码（最常用）

账号信息只从部署环境读取。Docker Compose 默认读取仓库根目录的 `.env`；该文件已被 Git 忽略，不应提交。

### 4.1 生成 bcrypt 密码哈希

在仓库根目录执行：

```bash
go run ./tools/hash_password
```

在标准输入中输入新密码并回车。工具要求至少 12 个字符，并输出 bcrypt cost 10 的哈希。

### 4.2 更新环境文件并重建后端容器

编辑根目录 `.env`：

```dotenv
CMS_AUTH_USERNAME=your-new-admin-name
CMS_AUTH_PASSWORD_HASH='$2a$10$replace-with-generated-hash'
CMS_AUTH_JWT_SECRET=replace-with-at-least-32-random-bytes
```

bcrypt 哈希必须保留单引号，避免其中的 `$` 被 Compose 展开。校验并重建后端容器：

```bash
docker compose config --quiet
docker compose up -d --force-recreate cms-api
```

不再接受 `CMS_AUTH_PASSWORD` 明文密码回退，也不会从 YAML 读取账号和密钥。

### 4.3 轮换 JWT 密钥

生成至少 32 字节的随机密钥：

```bash
openssl rand -hex 32
```

将输出写入 `.env` 的 `CMS_AUTH_JWT_SECRET`，然后重建后端容器：

```bash
docker compose up -d --force-recreate cms-api
```

> 修改 JWT 密钥会使旧 token 全部失效，用户需要重新登录。这是预期行为。

### 4.4 验证

```bash
curl -i http://127.0.0.1/api/health
docker compose logs --tail=100 cms-api
```

---

## 5. 如何查看 MySQL 数据

进入容器：

```bash
docker compose exec mysql mysql -uroot -proot
```

然后：

```sql
USE cms;
SHOW TABLES;
SELECT * FROM students LIMIT 20;
SELECT * FROM score_entries ORDER BY id DESC LIMIT 50;
```

升级后首次启动会自动核对 `students.total_score` 与 `score_entries` 汇总值。若存在旧数据差额，会新增“历史总分结转”明细以保留原总分。部署前应先完成数据库备份；启动后可用以下 SQL 抽查是否仍有不一致：

```sql
SELECT s.student_no, s.name, s.total_score, COALESCE(SUM(e.score), 0) AS entry_total
FROM students s
LEFT JOIN score_entries e ON e.student_id = s.id
GROUP BY s.id
HAVING s.total_score <> entry_total;
```

---

## 6. 备份与恢复（强烈建议做）

### 6.1 备份（逻辑备份，推荐）

在服务器上执行：

```bash
mkdir -p backups

docker compose exec -T mysql \
  mysqldump -uroot -proot --databases cms --single-transaction --routines --events \
  > backups/cms_$(date +%F_%H%M%S).sql
```

（可选）压缩：

```bash
gzip backups/cms_*.sql
```

把 `backups/` 同步到对象存储/另一台机器。

### 6.2 恢复（到当前环境）

```bash
cat backups/cms_xxx.sql | docker compose exec -T mysql mysql -uroot -proot
```

> 恢复会覆盖同名库里的表与数据，建议在维护窗口操作。

---

## 7. 常见问题排查

### 7.1 外网打不开网页

检查三件事：

1) 安全组/防火墙是否放行 80
2) 容器是否在跑：`docker compose ps`
3) 本机能否访问：`curl -I http://127.0.0.1/`

### 7.2 能打开前端但登录报错

1) `curl -i http://127.0.0.1/api/health` 是否 200
2) `docker compose logs --tail=200 cms-api` 查看错误
3) 执行 `docker compose config --quiet`，并确认 `.env` 中三个 `CMS_AUTH_*` 变量均已设置

### 7.3 数据丢失怀疑

确认卷是否还在：

```bash
docker volume ls | grep mysql_data
```

如果你执行过 `docker compose down -v`，数据卷会被删除，数据无法恢复（除非你有备份）。

积分明细没有按天数自动清理机制；若记录减少，应优先核对是否发生了主动撤销、数据卷删除或数据库恢复操作。
