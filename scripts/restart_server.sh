#!/bin/bash

# Marketing Service 重启脚本
# 用于快速重启服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 服务配置
SERVICE_NAME="marketing-service"
HTTP_PORT=8105
GRPC_PORT=9105
CONFIG_FILE="configs/config.yaml"
BIN_DIR="bin"
BIN_NAME="server"

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Marketing Service 重启脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 1. 停止正在运行的服务
echo -e "${YELLOW}[1/4] 停止正在运行的服务...${NC}"

# 查找占用 HTTP 端口的进程
HTTP_PID=$(lsof -ti:$HTTP_PORT 2>/dev/null || true)
# 查找占用 gRPC 端口的进程
GRPC_PID=$(lsof -ti:$GRPC_PORT 2>/dev/null || true)

# 合并所有 PID
ALL_PIDS=$(echo "$HTTP_PID $GRPC_PID" | tr ' ' '\n' | sort -u | grep -v '^$' || true)

if [ -n "$ALL_PIDS" ]; then
    echo "找到运行中的进程，正在停止..."
    for PID in $ALL_PIDS; do
        if ps -p $PID > /dev/null 2>&1; then
            echo "  停止进程 (PID: $PID)"
            kill -9 $PID 2>/dev/null || true
        fi
    done
    sleep 1
    echo -e "${GREEN}✓ 服务已停止${NC}"
else
    echo -e "${GREEN}✓ 没有运行中的服务${NC}"
fi
echo ""

# 2. 清理旧的编译文件
echo -e "${YELLOW}[2/4] 清理旧的编译文件...${NC}"
rm -f "$BIN_DIR/$BIN_NAME" 2>/dev/null || true
echo -e "${GREEN}✓ 清理完成${NC}"
echo ""

# 3. 编译服务
echo -e "${YELLOW}[3/4] 编译服务...${NC}"
if make build > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 编译成功${NC}"
else
    echo -e "${RED}✗ 编译失败，请检查错误信息${NC}"
    echo -e "${YELLOW}运行 'make build' 查看详细错误${NC}"
    exit 1
fi
echo ""

# 4. 启动服务
echo -e "${YELLOW}[4/4] 启动服务...${NC}"

# 检查可执行文件
if [ ! -f "$BIN_DIR/$BIN_NAME" ]; then
    echo -e "${RED}✗ 可执行文件不存在: $BIN_DIR/$BIN_NAME${NC}"
    exit 1
fi

# 检查配置文件
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}✗ 配置文件不存在: $CONFIG_FILE${NC}"
    exit 1
fi

# 确保 logs 目录存在
mkdir -p logs

# 启动服务（后台运行）
echo "  启动服务..."
nohup "$BIN_DIR/$BIN_NAME" -conf "$CONFIG_FILE" > logs/server.log 2>&1 &
NEW_PID=$!

# 等待服务启动
sleep 2

# 检查服务是否启动成功
if ps -p $NEW_PID > /dev/null 2>&1; then
    # 再次检查端口是否被占用（确认服务真正启动）
    sleep 1
    if lsof -ti:$HTTP_PORT > /dev/null 2>&1 || lsof -ti:$GRPC_PORT > /dev/null 2>&1; then
        echo -e "${GREEN}✓ 服务启动成功 (PID: $NEW_PID)${NC}"
        echo ""
        echo -e "${GREEN}========================================${NC}"
        echo -e "${GREEN}服务信息：${NC}"
        echo -e "  服务名称: $SERVICE_NAME"
        echo -e "  HTTP 端口: $HTTP_PORT"
        echo -e "  gRPC 端口: $GRPC_PORT"
        echo -e "  进程 ID: $NEW_PID"
        echo -e "  日志文件: logs/server.log"
        echo -e "${GREEN}========================================${NC}"
        echo ""
        echo -e "${YELLOW}查看日志: tail -f logs/server.log${NC}"
        echo -e "${YELLOW}停止服务: kill $NEW_PID${NC}"
    else
        echo -e "${RED}✗ 服务启动失败（端口未监听）${NC}"
        echo -e "${YELLOW}查看日志: tail -20 logs/server.log${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ 服务启动失败（进程不存在）${NC}"
    echo -e "${YELLOW}查看日志: tail -20 logs/server.log${NC}"
    exit 1
fi

