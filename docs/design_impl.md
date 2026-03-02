# 设计与实现说明（后端）

## 1. 技术栈

- Golang
- go-zero（REST 服务与路由）
- GORM（MySQL ORM）
- MySQL 8

## 2. 代码结构（关键目录）

- `backend/cms.go`：服务入口，加载配置、注册路由、AutoMigrate、启动定时清理
- `backend/internal/svc/servicecontext.go`：依赖注入（DB、各 Repo、RollcallState、LockRepo）
- `backend/internal/repository/*`：数据访问层
- `backend/internal/logic/*`：业务逻辑层
- `backend/internal/handler/*`：HTTP handler 层

## 3. 鉴权

- 登录：`POST /api/auth/login`，校验用户名与 bcrypt 密码哈希
- 受保护接口：除 `/api/health`、`/api/auth/login` 外，全部需要 `Authorization: Bearer <JWT>`
- JWT 解析强制校验签名算法为 HS256

## 4. 30 天明细保留（Retention）

- `retention.Runner` 使用 `time.Ticker` 周期触发清理
- `retention.Cleaner` 调用 `ScoreEntryRepo.DeleteBefore` 删除早于阈值的明细
- 多实例安全：Runner 通过 MySQL advisory lock（`GET_LOCK`）确保同一时刻只有一个实例执行清理

相关文件：

- `backend/internal/retention/runner.go`
- `backend/internal/retention/retention.go`
- `backend/internal/repository/lock_repo.go`

## 5. 学生 Excel 导入

- 接口：`POST /api/students/import`（multipart `file`）
- 校验：仅 `.xlsx`，最大 10MB；校验表头 A/B 列
- 数据写入：批量 upsert（按 `student_no` 冲突更新 name/gender/phone/position），并使用事务保证原子性

相关文件：

- `backend/internal/logic/studentimportlogic.go`
- `backend/internal/repository/student_repo.go`
