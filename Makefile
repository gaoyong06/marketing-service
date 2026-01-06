# Marketing Service Makefile (devops-tools integrated)

.PHONY: all

SERVICE_NAME=marketing-service
SERVICE_DISPLAY_NAME=Marketing Service
HTTP_PORT=8105
GRPC_PORT=9105
API_PROTO_DIR=api/marketing_service/v1
API_PROTO_PATH=api/marketing_service/v1/marketing.proto
WIRE_DIRS=cmd/server
TEST_CONFIG=test/api/api-test-config.yaml
RUN_MODE=debug

DEVOPS_TOOLS_DIR := $(shell cd .. && pwd)/devops-tools
include $(DEVOPS_TOOLS_DIR)/Makefile.common
