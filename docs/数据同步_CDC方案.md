# 数据同步（CDC / 异步日志）方案与模板

本文提供两种可选方案：Debezium 与 Maxwell，并给出 MQ/日志流的落地模板。用于跨服务数据同步与事件驱动。

---

## 方案 A：Debezium + Kafka（推荐）

### 1. 架构
```
MySQL -> Debezium Connector -> Kafka Topic -> Consumer -> 下游服务/ES/仓储
```

### 2. 配置模板（MySQL Connector）
```json
{
  "name": "mysql-connector",
  "config": {
    "connector.class": "io.debezium.connector.mysql.MySqlConnector",
    "database.hostname": "mysql",
    "database.port": "3306",
    "database.user": "root",
    "database.password": "password",
    "database.server.id": "184054",
    "database.server.name": "go-micro",
    "database.include.list": "go_micro",
    "table.include.list": "go_micro.orders,go_micro.inventory,go_micro.payments",
    "snapshot.mode": "initial",
    "include.schema.changes": "false",
    "topic.prefix": "cdc",
    "tombstones.on.delete": "false"
  }
}
```

### 3. 消费示例（伪代码）
```
topic: cdc.go_micro.orders
payload: { before, after, op, ts_ms }
```

### 4. 优点
- 支持结构化变更事件
- 可恢复、支持断点续传
- 适用于高吞吐场景

---

## 方案 B：Maxwell + RabbitMQ/Kafka

### 1. 架构
```
MySQL -> Maxwell -> MQ Topic -> Consumer -> 下游服务/ES
```

### 2. 配置模板（Maxwell）
```
producer=kafka
kafka.bootstrap.servers=localhost:9092
host=mysql
user=root
password=password
database=go_micro
table=orders,inventory,payments
```

### 3. 优点
- 部署轻量
- 输出 JSON 事件，易消费

---

## 异步日志流方案（ELK/EFK）

### 1. 架构
```
应用日志 -> Fluentd/Logstash -> Elasticsearch -> Kibana
```

### 2. Fluentd 模板（示例）
```
<source>
  @type tail
  path /var/log/app/*.log
  pos_file /var/log/td-agent/app.pos
  tag go-micro.*
  format json
</source>

<match go-micro.**>
  @type elasticsearch
  host elasticsearch
  port 9200
  logstash_format true
</match>
```

---

## 一致性与幂等规范
- 消费端必须幂等（建议使用业务唯一键去重）
- 事件需携带版本号或时间戳
- 失败重试需有最大重试与死信策略
