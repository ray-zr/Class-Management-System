# 班主任使用说明（API 版）

> 当前仓库以 Go 后端 API 为主，同时提供 `docs/frontend/` 下的轻量前端页面（由 Nginx 静态托管）。也可通过 Postman / curl 调用。

## 1. 登录

请求：`POST /api/auth/login`

```bash
curl -sS -H 'Content-Type: application/json' \
  -d '{"username":"teacher","password":"teacher"}' \
  http://127.0.0.1/api/auth/login
```

返回 `accessToken`，后续请求在 Header 加：

`Authorization: Bearer <token>`

## 2. 学生管理

### 2.1 列表

`GET /api/students`

```bash
curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1/api/students
```

### 2.2 Excel 导入

`POST /api/students/import`

- Content-Type：`multipart/form-data`
- 文件字段名：`file`
- 仅支持 `.xlsx`
- 文件大小上限：10MB
- 第一行必须是表头：
  - A 列：`StudentNo` 或 `学号`
  - B 列：`Name` 或 `姓名`
- 列定义：A 学号、B 姓名、C 性别、D 联系方式、E 班委职位

示例：

```bash
curl -H "Authorization: Bearer $TOKEN" \
  -F "file=@students.xlsx" \
  http://127.0.0.1/api/students/import
```

导入行为：

- 按学号 `student_no` upsert（已存在则更新姓名/性别/电话/职位）
- 发现任意行缺少学号或姓名会返回 400，并给出行号错误信息

## 3. 随机点名（公平模式）

- `POST /api/rollcall/start`
- `POST /api/rollcall/pick`
- `POST /api/rollcall/reset`

（具体请求体以 `docs/cms.api` 为准。）

## 4. 积分录入

- `POST /api/score-entries` 录入积分
- `GET /api/score-entries` 查询最近 30 天明细

系统会在后台定时清理超过 30 天的明细，但不影响总分统计。

## 5. 维度与积分项（新增 / 编辑 / 删除）

### 5.1 维度

- `POST /api/dimensions` 新增
- `GET /api/dimensions` 列表
- `PUT /api/dimensions/:id` 修改名称
- `DELETE /api/dimensions/:id` 删除（若该维度下存在积分项或已产生积分明细，会返回 400 阻止删除）

示例：

```bash
curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"课堂表现"}' \
  http://127.0.0.1/api/dimensions

curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"课堂表现（更新）"}' \
  http://127.0.0.1/api/dimensions/1

curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1/api/dimensions
curl -H "Authorization: Bearer $TOKEN" -X DELETE http://127.0.0.1/api/dimensions/1
```

### 5.2 积分项

- `POST /api/score-items` 新增
- `GET /api/score-items?dimensionId=...` 列表
- `PUT /api/score-items/:id` 修改
- `DELETE /api/score-items/:id` 删除（若该积分项已产生积分明细，会返回 400 阻止删除）

示例：

```bash
curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"dimensionId":1,"name":"积极发言","score":2}' \
  http://127.0.0.1/api/score-items

curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"dimensionId":1,"name":"积极发言（更新）","score":3}' \
  http://127.0.0.1/api/score-items/1

curl -H "Authorization: Bearer $TOKEN" "http://127.0.0.1/api/score-items?dimensionId=1"
curl -H "Authorization: Bearer $TOKEN" -X DELETE http://127.0.0.1/api/score-items/1
```
