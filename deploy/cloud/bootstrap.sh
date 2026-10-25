#!/usr/bin/env bash
# 云主机一键部署 R5 演示栈（Ubuntu/Debian）：
#   curl -fsSL <raw-url>/deploy/cloud/bootstrap.sh | bash
# 或 git clone 后执行： ./deploy/cloud/bootstrap.sh
set -euo pipefail

REPO_URL="${REPO_URL:-https://github.com/Humphrey-He/go-micro.git}"
BRANCH="${BRANCH:-main}"

command -v docker >/dev/null 2>&1 || {
  echo "== 安装 Docker =="
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker
}
docker compose version >/dev/null 2>&1 || {
  echo "== docker compose 插件缺失，请升级 Docker =="
  exit 1
}

WORKDIR="${WORKDIR:-$HOME/ecomm-demo}"
if [ -d "$WORKDIR/.git" ]; then
  echo "== 更新仓库 $WORKDIR =="
  git -C "$WORKDIR" fetch origin "$BRANCH"
  git -C "$WORKDIR" reset --hard "origin/$BRANCH"
else
  echo "== 克隆仓库到 $WORKDIR =="
  git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$WORKDIR"
fi

cd "$WORKDIR"
echo "== 拉起演示栈 =="
bash manage.sh up

echo
echo "=============================================="
echo " R5 演示栈已上线："
echo "   distribution : $(hostname -I 2>/dev/null | awk '{print $1}'):8090"
echo "   mock 平台     : :9091 / :9092 / :9093"
echo " 管理命令（在 $WORKDIR 下）："
echo "   ./manage.sh status | demo | demo-loss | logs | down | clean"
echo " 验收演示：./manage.sh demo 50  （零 429 断言）"
echo "=============================================="
