# test-formatter

一个使用 Go 实现的简单 RESTful 参数格式化服务。

## 功能

- `POST /param_formatter`
  - 根据输入 `backup_type` 返回对应的备份参数模板。
  - 支持两种类型：`db`、`vm`。
  - 返回结构为三层嵌套，参数字段总数超过 20。
- `GET /schema`
  - 返回接口 schema（包含请求与响应结构说明）。

## 本地运行

```bash
go run main.go
```

服务默认监听：`http://127.0.0.1:8080`

## 接口示例

### 1) 获取 DB 备份参数模板

```bash
curl -X POST http://127.0.0.1:8080/param_formatter \
  -H 'Content-Type: application/json' \
  -d '{"backup_type":"db"}'
```

### 2) 获取 VM 备份参数模板

```bash
curl -X POST http://127.0.0.1:8080/param_formatter \
  -H 'Content-Type: application/json' \
  -d '{"backup_type":"vm"}'
```

### 3) 获取 schema

```bash
curl http://127.0.0.1:8080/schema
```

## Docker 运行

### 方式一：使用脚本一键运行

```bash
./scripts/docker-run.sh
```

脚本会自动：
1. 构建镜像 `param-formatter:local`
2. 删除旧容器（如存在）
3. 启动新容器并映射端口 `8080:8080`

### 方式二：手动运行

```bash
# 构建镜像
docker build -t param-formatter:local .

# 启动容器
docker run -d --name param-formatter -p 8080:8080 param-formatter:local
```

## 停止容器

```bash
docker rm -f param-formatter
```
