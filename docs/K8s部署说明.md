# Kubernetes 部署说明

本文档说明如何在 Kubernetes 环境中部署 go-micro 微服务。

## 前提条件

- Kubernetes 1.24+
- kubectl 已配置
- Docker 镜像已推送到 registry
- Secrets 已配置（KUBECONFIG_STAGING, KUBECONFIG_PROD）

## 目录结构

```
deploy/k8s/
├── env/                    # 环境配置
│   ├── dev.yaml           # 开发环境
│   ├── staging.yaml       # 预发布环境
│   └── prod.yaml          # 生产环境
├── common/                # 通用资源
│   ├── serviceaccount.yaml
│   ├── networkpolicy.yaml
│   └── poddisruptionbudget.yaml
├── gateway-api.yaml
├── order-service.yaml
├── payment-service.yaml
├── inventory-service.yaml
├── user-service.yaml
├── task-service.yaml
├── activity-service.yaml
├── price-service.yaml
└── refund-service.yaml
```

## 部署步骤

### 1. 创建命名空间

```bash
# 开发环境
kubectl create namespace go-micro-dev

# 预发布环境
kubectl create namespace go-micro-staging

# 生产环境
kubectl create namespace go-micro-prod
```

### 2. 应用环境配置

```bash
# 开发环境
kubectl apply -f deploy/k8s/env/dev.yaml

# 预发布环境
kubectl apply -f deploy/k8s/env/staging.yaml

# 生产环境
kubectl apply -f deploy/k8s/env/prod.yaml
```

### 3. 应用通用资源

```bash
kubectl apply -f deploy/k8s/common/
```

这将创建：
- ServiceAccount (`go-micro-sa`)
- RBAC Role 和 RoleBinding
- NetworkPolicy
- PodDisruptionBudget

### 4. 部署服务

```bash
# 替换镜像地址后部署
IMG=your-registry/gateway-api:latest

# 部署 gateway
sed "s|your-registry/gateway-api:latest|${IMG}|g" deploy/k8s/gateway-api.yaml | kubectl apply -f -

# 部署其他服务类似...
```

## GitHub Actions 部署

### Staging 自动部署

main 分支推送后自动触发：

```yaml
# .github/workflows/deploy-staging.yml
on:
  push:
    branches: [ "main" ]
```

### Production 手动部署

通过 workflow_dispatch 手动触发：

```bash
# 需要提供
# - source_tag: 镜像标签
# - reason: 部署原因
```

### 回滚

通过 rollback workflow 回滚：

```bash
# 需要提供
# - environment: staging 或 prod
# - reason: 回滚原因
```

## 验证部署

### 检查 Pod 状态

```bash
kubectl get pods -n go-micro -o wide
```

### 检查服务日志

```bash
kubectl logs -n go-micro deployment/gateway-api -f
```

### 检查服务健康

```bash
kubectl exec -it -n go-micro deployment/gateway-api -- curl localhost:8080/health
```

## 配置说明

### 环境变量

| 变量 | 说明 | 示例 |
|-----|------|------|
| APP_ENV | 环境名称 | production |
| REDIS_ADDR | Redis 地址 | redis:6379 |
| MYSQL_MAX_OPEN_CONNS | MySQL 最大连接数 | 100 |
| RATE_LIMIT_QPS | QPS 限制 | 500 |

### 密钥

生产环境密钥必须通过 Secret 管理：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: go-micro-secret
type: Opaque
stringData:
  MYSQL_DSN: "prod_user:password@tcp(mysql:3306)/go_micro?..."
  JWT_SECRET: "your-32-char-secret-key"
```

## 扩缩容

### 手动扩缩容

```bash
kubectl scale deployment gateway-api -n go-micro --replicas=3
```

### 自动扩缩容 (HPA)

需要额外配置 HorizontalPodAutoscaler：

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gateway-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gateway-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

## 故障排查

### Pod 无法启动

```bash
kubectl describe pod <pod-name> -n go-micro
kubectl logs <pod-name> -n go-micro
```

### 服务无法访问

1. 检查 Service 是否正确配置
2. 检查 NetworkPolicy 是否阻止了流量
3. 检查 endpoints 是否存在

### 滚动更新卡住

```bash
kubectl rollout status deployment/<name> -n go-micro
kubectl rollout undo deployment/<name> -n go-micro
```

## 安全配置

### 当前安全措施

- 非 root 用户运行 (`runAsNonRoot: true`)
- 使用 seccomp 限制系统调用
- 禁用特权容器
- 只读根文件系统
- 丢弃所有 capabilities
- NetworkPolicy 限制网络访问
- ServiceAccount 最小权限原则

### 建议的生产增强

1. 启用 PodSecurityPolicy 或 Pod Security Standards
2. 配置 ResourceQuota 限制资源使用
3. 启用 AuditPolicy 记录操作日志
4. 使用 Vault 或 AWS Secrets Manager 管理密钥
