# test-formatter

一个使用 Go 实现的简单 RESTful 参数格式化服务。

## 功能

- `POST /param_formatter`
  - 输入完整备份参数对象，返回原样对象（echo）。
  - 输入和输出结构保持一致：`{ xxxx } -> { xxxx }`。
  - 三层嵌套，参数字段总数超过 20。
  - 支持两种资源类型校验：`db`、`vm`。
- `GET /schema`
  - 返回接口 schema（包含请求与响应结构说明）。

## 本地运行

```bash
go run main.go
```

服务默认监听：`http://127.0.0.1:8080`

## 接口示例

### 1) DB 参数 echo（输入即输出）

```bash
curl -X POST http://127.0.0.1:8080/param_formatter \
  -H 'Content-Type: application/json' \
  -d '{
    "job": {
      "job_id": "job-10001",
      "job_name": "nightly-backup",
      "job_type": "full",
      "priority": "normal",
      "tenant_id": "tenant-a",
      "operator_id": "user-ops-01",
      "tags": ["prod", "critical", "db"]
    },
    "source": {
      "resource": {
        "resource_type": "db",
        "resource_id": "db-001",
        "resource_name": "orders-mysql",
        "cluster_id": "",
        "namespace": "",
        "host": "10.0.0.21",
        "port": 3306,
        "database_name": "orders",
        "vm_uuid": "",
        "hypervisor": ""
      },
      "auth": {
        "credential_ref": "credential/default",
        "auth_mode": "token"
      }
    },
    "target": {
      "storage": {
        "provider": "s3",
        "bucket": "backup-bucket",
        "path": "/daily",
        "region": "us-east-1",
        "storage_class": "standard",
        "kms_key_id": "kms-key-001"
      },
      "retention": {
        "mode": "days",
        "keep_last": 30,
        "expire_after_days": 180
      }
    },
    "policy": {
      "schedule": {
        "enabled": true,
        "type": "cron",
        "cron_expr": "0 2 * * *",
        "timezone": "UTC",
        "start_at": "2026-04-01T02:00:00Z"
      },
      "consistency": {
        "app_consistent": true,
        "quiesce_fs": true,
        "pre_script_ref": "script/pre-freeze",
        "post_script_ref": "script/post-thaw"
      },
      "security": {
        "encrypt_in_transit": true,
        "encrypt_at_rest": true,
        "password_protected": false,
        "password_ref": ""
      }
    },
    "execution": {
      "retry": {
        "max_attempts": 3,
        "backoff_seconds": 15
      },
      "performance": {
        "bandwidth_limit_mbps": 500,
        "parallelism": 8,
        "dedup": true,
        "compression": "lz4"
      },
      "notification": {
        "on_success": true,
        "on_failure": true,
        "channel": "webhook",
        "recipient_ref": "notify/ops-webhook"
      }
    }
  }'
```

### 2) VM 参数 echo（输入即输出）

```bash
curl -X POST http://127.0.0.1:8080/param_formatter \
  -H 'Content-Type: application/json' \
  -d @vm_payload.json
```

其中 `vm_payload.json` 与上面的结构一致，仅将 `source.resource.resource_type` 设置为 `vm`，并按需填写 `cluster_id`、`vm_uuid`、`hypervisor`。

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
