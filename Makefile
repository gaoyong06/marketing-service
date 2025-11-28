# Marketing Service Makefile (devops-tools integrated)

.PHONY: all

SERVICE_NAME=marketing-service
SERVICE_DISPLAY_NAME=Marketing Service
HTTP_PORT=8001
GRPC_PORT=9001
API_PROTO_DIR=
API_PROTO_PATH=
TEST_CONFIG=

DEVOPS_TOOLS_DIR := $(shell cd .. && pwd)/devops-tools
include $(DEVOPS_TOOLS_DIR)/Makefile.common
