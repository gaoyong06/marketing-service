# Marketing Service Makefile

.PHONY: all build wire proto clean

# 默认任务：构建服务
all: proto wire build

# 构建服务
build:
	go build -o ./bin/ ./...

# 生成 wire 依赖注入代码
wire:
	cd cmd/marketing-service && wire

# 确保 proto-repo 中的代码已生成并更新依赖
proto:
	@echo "检查 proto-repo 中的代码..."
	@if [ ! -d "../proto-repo/gen/go/platform/marketing_service" ]; then \
		echo "proto-repo 中没有找到生成的代码，正在生成..."; \
		cd ../proto-repo && ./scripts/generate.sh; \
	fi
	@echo "更新 go.mod 中的依赖..."
	go mod tidy

# 清理生成的文件
clean:
	rm -rf bin/
	rm -rf api/platform/
	rm -rf api/base/
	@echo "清理完成！"
