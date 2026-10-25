#!/usr/bin/env bash
# R5 分发调度演示栈管理：上线 / 下线 / 状态 / 日志 / 彻底删除 / 验收演示
#
#   ./manage.sh up        构建并拉起整套（mysql/redis/3个mock平台/distribution）
#   ./manage.sh down      停止（保留数据卷）
#   ./manage.sh clean     彻底删除（容器+数据卷+本地镜像）
#   ./manage.sh status    各容器状态与健康检查
#   ./manage.sh logs [svc] 跟踪日志（默认 distribution-service）
#   ./manage.sh demo      验收演示：批量分发→平台统计→零429断言
#   ./manage.sh demo-loss 回执丢失演示：拨码丢回执→diff抓取→补发→台账
set -euo pipefail
cd "$(dirname "$0")"

COMPOSE="docker compose -f deploy/cloud/docker-compose.demo.yml -p ecomm-demo"
DIST=http://localhost:8090

case "${1:-help}" in
  up)
    $COMPOSE up -d --build
    echo "== 等待服务就绪 =="
    for i in $(seq 1 60); do
      if curl -sf ${DIST}/healthz > /dev/null 2>&1; then break; fi
      sleep 2
    done
    curl -sf ${DIST}/healthz && echo " <- distribution OK"
    echo "演示栈已上线: distribution=:8090 mock平台=:9091/9092/9093"
    ;;
  down)
    $COMPOSE down
    echo "已下线（数据卷保留，up 可恢复）"
    ;;
  clean)
    $COMPOSE down -v --rmi local
    echo "已彻底删除（容器+数据卷+本地镜像）"
    ;;
  status)
    $COMPOSE ps
    echo "--- health ---"
    curl -s ${DIST}/healthz || echo "distribution DOWN"
    for port in 9091 9092 9093; do
      curl -s http://localhost:${port}/healthz || echo " (mock:${port} DOWN)"
      echo
    done
    ;;
  logs)
    $COMPOSE logs -f --tail=100 "${2:-distribution-service}"
    ;;
  demo)
    N="${2:-50}"
    echo "== 1) 分发 ${N} 个 SKU（每个展开到 3 平台 = $((N*3)) 任务，店铺桶 15QPS 削峰）=="
    for i in $(seq 1 "${N}"); do
      SKU=$(printf "SKU-%04d" $((RANDOM % 1000 + 1)))
      curl -s -X POST ${DIST}/distribution/dispatch -H 'Content-Type: application/json' \
        -d "{\"sku_id\":\"${SKU}\"}" > /dev/null
    done
    echo "等待调度消化..."
    sleep 20
    echo "== 2) 任务状态分布 =="
    curl -s ${DIST}/distribution/task-counts
    echo
    echo "== 3) 三个 mock 平台统计（验收点：rejected_429 全部为 0）=="
    for port in 9091 9092 9093; do
      echo -n "platform:${port} -> "
      curl -s http://localhost:${port}/admin/stats
      echo
    done
    ;;
  demo-loss)
    echo "== 1) 平台 9093 拨码：回执丢失率 30% =="
    curl -s -X POST http://localhost:9093/admin/config -H 'Content-Type: application/json' -d '{"loss_rate":0.3}'
    echo
    echo "== 2) 分发 30 个 SKU =="
    for i in $(seq 1 30); do
      SKU=$(printf "SKU-%04d" $((RANDOM % 1000 + 1)))
      curl -s -X POST ${DIST}/distribution/dispatch -H 'Content-Type: application/json' \
        -d "{\"sku_id\":\"${SKU}\"}" > /dev/null
    done
    sleep 15
    echo "== 3) 触发一轮回执比对（生产为 15 分钟轮次，演示手动触发）=="
    curl -s -X POST ${DIST}/distribution/reconcile/tick
    echo
    sleep 10
    echo "== 4) 再跑一轮（补发后的任务应拿到回执 → CONFIRMED）=="
    curl -s -X POST ${DIST}/distribution/reconcile/tick
    echo
    echo "== 5) 差异台账（RECEIPT_MISSING 应有记录，状态机 PENDING→RESOLVED）=="
    mysql -h 127.0.0.1 -P 3307 -uroot -proot go_micro \
      -e "SELECT id, sku_id, request_id, status FROM reconcile_diff WHERE biz_type='DISTRIBUTION' ORDER BY id DESC LIMIT 10;" 2>/dev/null \
      || echo "(本机无 mysql 客户端，跳过台账打印——用 docker compose exec mysql mysql 进入查看)"
    echo "演示完成。恢复平台回执：curl -X POST http://localhost:9093/admin/config -d '{\"loss_rate\":0}'"
    ;;
  *)
    sed -n '2,14p' "$0"
    ;;
esac
