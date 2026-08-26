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

- 通过网页端：在“学生管理”页面点击“下载导入模板”
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
- `DELETE /api/score-entries/:id` 撤销一条积分明细；请求体必须提供 `reason`（会回滚该条记录对学生总分的影响）
- `GET /api/operation-logs` 查询加减分操作日志，可用 `startDate`、`endDate` 按日期范围筛选

网页端“积分录入”页面将加分项和扣分项分列展示，可按维度筛选，也可输入积分项名称或维度关键字进行模糊筛选。

积分明细不会自动清理。主动撤销后记录仍会保留，并显示撤销时间和原因，但不再参与学生总分、排行榜或导出统计。网页端通过侧边栏“积分管理”区域的“积分记录日志”进入，默认查询当月，可按开始和结束日期筛选。

系统升级启动时会核对每名学生的当前总分与全部积分明细之和。若旧数据存在差额，系统不会删除总分或伪造具体班规行为，而是新增一条“历史数据校准 / 历史总分结转”明细，分值为差额，并保存学生、组、维度和积分细则名称快照。校准后，全部明细相加等于当前总分；该过程幂等，重复启动不会重复结转。

网页端会自动为每次计分生成请求 ID；网络失败后重试同一操作不会重复加分。通过 API 调用时也可传入不超过 64 字节的 `requestId` 获得相同保护。

删除学生后，该学生不会出现在当前名单、当前总分榜或点名池中，但历史积分仍可查看和撤销。操作日志显示录入当时的学生、组、维度和积分项名称。

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

### 5.3 按维度查看名单

网页端进入“积分规则”，点击任一维度右侧的“统计”，选择开始和结束日期后可查看：

- 有加分：该时段在该维度至少有一条正分记录的当前学生，并显示加分合计。
- 有减分：该时段在该维度至少有一条负分记录的当前学生，并显示扣分合计；同时有正、负记录的学生会出现在两个名单中。
- 无加减分：该时段在该维度没有任何积分记录的当前学生。

## 6. 导出班级量化汇总和细则

网页端进入“积分排行榜”，选择开始日期、结束日期和可选维度，点击“导出量化汇总及细则”。勾选“总分榜”时导出全部历史。

导出的 Excel 包含两个工作表：

- `量化汇总`：按学号自然顺序、姓名排序，包含各维度净分、加分合计、扣分合计、期间净分和当前总分。
- `积分细则`：按学号、姓名、记录时间排序，包含每条记录的时间、维度快照、积分细则快照、分值和备注。
