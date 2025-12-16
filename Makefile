# Public area
TOOLS_MOD_DIR := ./tools
PROTO_GEN_DIR := ./genproto
GO_MOD_CACHE_DIR := ./.gocache
ALL_GO_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | grep -v $(GO_MOD_CACHE_DIR) | sort | uniq) # find all go.mod files
ROOT_GO_MOD_DIRS := $(filter-out $(TOOLS_MOD_DIR) $(PROTO_GEN_DIR) $(GO_MOD_CACHE_DIR), $(ALL_GO_MOD_DIRS)) # filter out tools directory
ALL_WIRE_DIRS := $(shell find . -type f -name 'wire.go' -exec dirname {} \; | sort | uniq) # find all wire.go files
# ALL_PROJECT_DIRS := $(shell find ./app -maxdepth 1 -mindepth 1 -type d | sort | uniq) # find all project directories
ALL_BUF_DIRS := $(shell find . -type f -name 'buf.yaml' -exec dirname {} \; | grep -v $(GO_MOD_CACHE_DIR) | sort | uniq) # find all buf.yaml files
ALL_AIR_DIRS := $(shell find . -type f -name '.air.toml' -exec dirname {} \; | sort | uniq) # find all air.toml files

GO = go
GIT = git
TIMEOUT = 60
TOOLS = $(CURDIR)/.tools
BIN_DIR = $(CURDIR)/bin
TMP_DIR = $(CURDIR)/tmp

# Gitlab - 延迟求值，只在需要时执行脚本
DIFF_RANGE = $(shell ./tools/scripts/changed_diff_args.sh)

# 禁用子目录的递归输出
MAKEFLAGS += --no-print-directory

# ==================================

# 自动检测是否运行在 CI 环境，禁用颜色输出
COLOR_ENABLE := true
ifneq (,$(findstring true,$(CI)))
  COLOR_ENABLE := false
endif

# Define Echo
# 彩色输出函数
# $(1): 颜色代码 (例如: 1;36 表示亮青色)
# $(2): 输出内容
# 用法: $(call echo_color, 1;36, "Hello, World!")

define echo_color
@if [ "$(COLOR_ENABLE)" = "true" ]; then \
    echo "\033[$(1)m$(2)\033[0m"; \
  else \
    echo "$(2)"; \
fi
endef

# 预定义的彩色输出函数
# 绿色成功
# Usage: $(call echo_success, "Success message")
define echo_success
$(call echo_color,1;32,$(1))
endef

# 红色错误
# Usage: $(call echo_error, "Error message")
define echo_error
$(call echo_color,1;31,$(1))
endef

# 黄色警告
# Usage: $(call echo_warning, "Warning message")
define echo_warning
$(call echo_color,1;33,$(1))
endef

# 蓝色信息
# Usage: $(call echo_info, "Info message")
define echo_info
$(call echo_color,1;36,$(1))
endef

# 紫色提示
# Usage: $(call echo_note, "Note message")
define echo_note
$(call echo_color,1;35,$(1))
endef



# ==================================

# Print
.PHONY: print-env print-diff
print-env:
	@echo "=================================="
	@echo " Environment Info:"
	@echo "=================================="
	@echo " Go Version:    $$(go version)"
	@echo " Go Env:        $$(go env | grep GOPATH)"
	@echo " Root Dir:      $(CURDIR)"
	@echo " Tools Dir:     $(TOOLS)"
	@echo " Bin Dir:       $(BIN_DIR)"
	@echo " Tmp Dir:       $(TMP_DIR)"
	@echo " Go Modules:    $(ALL_GO_MOD_DIRS)"
# 	@echo " Projects:      $(ALL_PROJECT_DIRS)"
	@echo " Buf Dirs:      $(ALL_BUF_DIRS)"
	@echo " Air Dirs:      $(ALL_AIR_DIRS)"
	@echo "=================================="

print-diff:
	@echo "=================================="
	@echo " Changes compared to $(DIFF_RANGE):"
	@echo "=================================="
	@CHANGES=$$($(GIT) diff --name-only $(DIFF_RANGE) || echo ""); \
  	if [ -z "$$CHANGES" ]; then \
		echo " No changes detected."; \
	else \
		echo " $$CHANGES" | while read -r line; do \
			echo " - $$line"; \
		done; \
	fi
	@echo "=================================="


# ==================================
# Lint

# 格式化所有子项目的 go.mod 文件
.PHONY: go-mod-tidy
go-mod-tidy: $(ALL_GO_MOD_DIRS:%=go-mod-tidy/%)
go-mod-tidy/%: DIR=$*
go-mod-tidy/%:
	$(call echo_info,"$(GO) mod tidy in $(DIR)") \
		&& cd $(DIR) \
		&& $(GO) mod tidy -compat=1.25.0

# 检查所有子项目的 go.mod 文件是否有变化
.PHONY: go-mod-tidy-diff
go-mod-tidy-diff: $(ALL_GO_MOD_DIRS:%=go-mod-tidy-diff/%)
go-mod-tidy-diff/%: DIR=$*
go-mod-tidy-diff/%:
	@set -e; \
		if git diff --name-only $(DIFF_RANGE) -- $(DIR) | grep -qE '\.(go|mod|sum)$$'; then \
			echo "🔧 go mod tidy in $(DIR)" && $(MAKE) go-mod-tidy/$(DIR); \
		fi

# Usage:
# make go-mod-update                                    			# 更新所有子项目的所有包（谨慎使用）
# make go-mod-update PACKAGE=github.com/flc1125/go-cron 			# 更新所有子项目中指定包
# make go-mod-update PACKAGE=github.com/flc1125/go-cron/... 		# 更新所有子项目中指定包及其子包
# make go-mod-update/app/dir                           				# 更新特定子项目的所有包
# make go-mod-update/app/dir PACKAGE=github.com/flc1125/go-cron 	# 更新特定子项目中的指定包
# make go-mod-update/app/dir PACKAGE=github.com/flc1125/go-cron/... # 更新特定子项目中的指定包及其子包
.PHONY: go-mod-update
go-mod-update: $(ALL_GO_MOD_DIRS:%=go-mod-update/%)
go-mod-update/%: DIR=$*
go-mod-update/%:
	$(call echo_info,"$(GO) mod update in $(DIR)") \
		&& cd $(DIR) \
		&& if [ -z "$(PACKAGE)" ] || grep -q "$(shell echo $(PACKAGE) | sed 's/@.*//')" go.mod; then \
		  	echo "😄update: $(DIR) need package $(PACKAGE)"; \
			$(GO) get -u $(if $(PACKAGE),$(PACKAGE),./...); \
		else \
		  	echo "😐skip: $(DIR) does not need package $(PACKAGE)"; \
		fi

# Usage:
# make go-mod-list         # 列出所有子项目的依赖包
# make go-mod-list/app/dir # 列出特定子项目的依赖包
.PHONY: go-mod-list
go-mod-list: $(ALL_GO_MOD_DIRS:%=go-mod-list/%)
go-mod-list/%: DIR=$*
go-mod-list/%:
	$(call echo_info,"$(GO) list -m all in $(DIR)") \
		&& cd $(DIR) \
		&& $(GO) list -m all

# Usage:
# make test         # 列出所有子项目的依赖包
# make test/app/dir # 列出特定子项目的依赖包
.PHONY: test
test: $(ALL_GO_MOD_DIRS:%=test/%)
test/%: DIR=$*
test/%:
	$(call echo_info,"$(GO) test in $(DIR)") \
		&& cd $(DIR) \
		&& $(GO) test ./...


# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN { commentBuffer = "" } \
	/^#/ { \
		# 收集注释行，用换行符分隔 \
		if (commentBuffer == "") { \
			commentBuffer = substr($$0, 3); \
		} else { \
			commentBuffer = commentBuffer "\n" substr($$0, 3); \
		} \
		next; \
	} \
	/^\.PHONY:/ { \
		# 遇到 .PHONY 行，保持注释缓冲区不变 \
		next; \
	} \
	/^[a-zA-Z\-_0-9]+:/ { \
		# 遇到目标行，如果有注释缓冲区就使用它 \
		if (commentBuffer != "") { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			# 处理多行注释的缩进 \
			split(commentBuffer, commentLines, "\n"); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand, commentLines[1]; \
			for (i = 2; i <= length(commentLines); i++) { \
				if (commentLines[i] != "") { \
					printf "%22s %s\n", "", commentLines[i]; \
				} \
			} \
			commentBuffer = ""; \
		} \
	} \
	!/^#/ && !/^\.PHONY:/ && !/^[a-zA-Z\-_0-9]+:/ { \
		# 遇到其他非注释、非.PHONY、非目标行，清空注释缓冲区 \
		commentBuffer = ""; \
	}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
