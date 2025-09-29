APP          := ada-cli
PKG          := ./cmd/$(APP)
ARTIFACTS    := artifacts
BUILD_DIR    := $(ARTIFACTS)/builds
CONF_SRC     := cfg
PRMPT_SRC    := prompts
DBG_SRC      := diagnostics
CONF_DST     := $(BUILD_DIR)/$(CONF_SRC)
PRMPT_DST    := $(BUILD_DIR)/$(PRMPT_SRC)
DBG_DST      := $(BUILD_DIR)/$(DBG_SRC)
stage        := 0

ifeq ($(OS),Windows_NT)
  EXE := .exe
  MKDIR = if not exist "$(1)" mkdir "$(1)"
  RMDIR = if exist "$(1)" rmdir /S /Q "$(1)"
  COPY  = if exist "$(1)" xcopy /E /I /Y "$(1)" "$(2)" >nul
else
  EXE :=
  MKDIR = mkdir -p $(1)
  RMDIR = rm -rf $(1)
  COPY  = cp -r $(1) $(2)
endif

EXECUTABLE := $(BUILD_DIR)/$(APP)$(EXE)

.PHONY: all build copy run clean install

.DEFAULT_GOAL := all

all: build copy

build:
	@echo Building $(APP)...
	@$(call MKDIR,$(ARTIFACTS))
	@$(call MKDIR,$(BUILD_DIR))
	@go build -ldflags="-s -w" -o $(EXECUTABLE) $(PKG)

copy:
	@echo Copying data...
	@$(call RMDIR,$(CONF_DST))
	@$(call RMDIR,$(PRMPT_DST))
	@$(call COPY,$(CONF_SRC),$(CONF_DST))
	@$(call COPY,$(PRMPT_SRC),$(PRMPT_DST))

debug: build copy
	@echo Running [flags: -debug, -stage=$(stage)] $(EXECUTABLE)...
	@$(call RMDIR,$(DBG_DST))
	@$(call COPY,$(DBG_SRC),$(DBG_DST))
	@$(EXECUTABLE) -debug -stage=$(stage)

run:
	@echo Running $(EXECUTABLE)...
	@$(EXECUTABLE)

clean:
	@echo Cleaning...
	@$(call RMDIR,$(BUILD_DIR))
	@go clean

br: build run
bcd: build copy debug
