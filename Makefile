OS := $(shell uname -s)
COMPOSE ?= docker compose
SLEEP_SECONDS ?= 20
SUMMARY_FILES := stress/summary-ip.html stress/summary-token.html
LYNX_FLAGS ?= -dump -assume_charset=utf-8 -display_charset=utf-8

all: ## Run full local flow (up + install-lynx + wait-stress + read-files)
	@$(MAKE) up install-lynx wait-stress read-files

up: ## Start services with Docker Compose
	@echo "📦 Starting services with Docker Compose..."
	@$(COMPOSE) up -d
	@echo "✅ Services are up."

down: ## Stop services with Docker Compose
	@echo "🛑 Stopping services with Docker Compose..."
	@$(COMPOSE) down
	@echo "✅ Services are down."

restart: ## Restart services with Docker Compose
	@echo "🔄 Restarting services with Docker Compose..."
	@$(MAKE) down up
	@echo "✅ Services restarted."

help: ## Display this help message
	@echo "📜 Makefile Help:"
	@awk 'BEGIN {FS = ":.*## "; printf "\n"} /^[a-zA-Z0-9_.-]+:.*## / { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

install-lynx: ## Install lynx for the current OS
	@echo "🔧 Installing lynx..."
	@if command -v lynx >/dev/null 2>&1; then \
		echo "Lynx already installed."; \
	elif [ "$(OS)" = "Darwin" ]; then \
		$(MAKE) install-lynx-macos; \
	elif [ -f /etc/redhat-release ] || [ -f /etc/debian_version ]; then \
		$(MAKE) install-lynx-linux; \
	else \
		echo "Sistema operacional não suportado."; \
		exit 1; \
	fi
	@echo "✅ Lynx instalado."

install-lynx-macos: ## (internal) Install lynx on macOS via Homebrew
	@echo "Detected macOS. Installing with Homebrew..."
	@if command -v brew >/dev/null 2>&1; then \
		brew install lynx; \
	else \
		echo "Homebrew not found. Install Homebrew first: https://brew.sh"; \
		exit 1; \
	fi

install-lynx-linux: ## (internal) Install lynx on Linux (dnf/apt)
	@if [ -f /etc/redhat-release ]; then \
		echo "Detectado Fedora/RHEL/CentOS. Instalando com dnf..."; \
		sudo dnf install -y lynx; \
	elif [ -f /etc/debian_version ]; then \
		echo "Detectado Debian/Ubuntu. Instalando com apt..."; \
		sudo apt update && sudo apt install -y lynx; \
	else \
		echo "Sistema Linux não suportado."; \
		exit 1; \
	fi

read-files: ## Read generated stress HTML reports with lynx
	@echo "📖 Lendo arquivos HTML com lynx..."
	@for file in $(SUMMARY_FILES); do \
		if [ -f "$$file" ]; then \
			lynx $(LYNX_FLAGS) "$$file" || true; \
		else \
			echo "⚠️  Arquivo não encontrado: $$file"; \
		fi; \
	done
	@echo "✅ Leitura concluída."

sleep: ## Wait before reading generated stress reports
	@echo "⏳ Aguardando geração de arquivos HTML para leitura com o Lynx ($(SLEEP_SECONDS) segundos)..."
	@sleep $(SLEEP_SECONDS)
	@echo "⏳ Tempo de espera concluído."

wait-stress: ## Wait until stress containers finish execution
	@echo "⏳ Aguardando término dos containers de stress..."
	@for service in stress-ip stress-token; do \
		cid=`$(COMPOSE) ps -q $$service`; \
		if [ -n "$$cid" ]; then \
			while [ "$$(docker inspect -f '{{.State.Running}}' $$cid 2>/dev/null)" = "true" ]; do \
				sleep 1; \
			done; \
		fi; \
	done
	@echo "✅ Containers de stress finalizados."

.PHONY: all up down restart help install-lynx install-lynx-macos install-lynx-linux read-files sleep wait-stress
