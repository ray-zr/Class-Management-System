# 班主任使用说明（API 版）

> 当前仓库以 Go 后端 API 为主，同时提供 `docs/frontend/` 下的轻量前端页面（由 Nginx 静态托管）。也可通过 Postman / curl 调用。

## 1. 登录

请求：`POST /api/auth/login`

```bash
curl -sS -H 'Content-Type: application/json' \
  -d '{"username":"your-admin-name","password":"your-password"}' \
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
- 每次最多 500 行学生数据
- 第一行必须是表头：
  - A 列：`StudentNo` 或 `学号`
  - B 列：`Name` 或 `姓名`
- 列定义：A 学号、B 姓名、C 性别、D 联系方式、E 班委职位

模板下载：

- 通过网页端：在“学生名单”页面点击“下载导入模板”
- 或直接下载：`/templates/students_import_template.xlsx`

示例：

```bash
curl -H "Authorization: Bearer $TOKEN" \
  -F "file=@students.xlsx" \
  http://127.0.0.1/api/students/import
```

导入行为：

- 按学号 `student_no` upsert（已存在则更新姓名/性别/电话/职位）
- 文件内学号不能重复；字段缺失、过长或重复会返回 400，并给出行号错误信息
- 任意一行校验失败时整次导入不写入
- 已软删除的学生使用相同学号导入时会恢复，并先放入未分组

## 3. 随机点名（公平模式）

- `POST /api/rollcall/start`
- `POST /api/rollcall/pick`
- `POST /api/rollcall/reset`

（具体请求体以 `docs/cms.api` 为准。）

## 4. 积分录入

- `POST /api/score-entries` 录入积分
- `GET /api/score-entries` 查询积分明细；省略 `sinceDays` 或传 `0` 查询全部
- `DELETE /api/score-entries/:id` 撤销一条积分明细（会回滚该条记录对学生总分的影响）

积分明细不会自动清理，除非主动撤销，否则永久保留。网页端默认筛选最近 30 天，可将“最近天数”设为 `0` 查看全部。

网页端会自动为每次计分生成请求 ID；网络失败后重试同一操作不会重复加分。通过 API 调用时也可传入不超过 64 字节的 `requestId` 获得相同保护。

删除学生后，该学生不会出现在当前名单、当前总分榜或点名池中，但历史积分仍可查看和撤销。历史明细显示录入当时的学生、组、维度和积分项名称。

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
