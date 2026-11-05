#!/bin/bash
set -e

# MCAP - Minecraft Server Management Panel
# One-command installer and launcher
# Supports: Ubuntu, Fedora, Debian, Arch, Windows (WSL), macOS

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "========================================"
echo -e "${BLUE}MCAP - One-Command Installer & Launcher${NC}"
echo "========================================"
echo ""

# Detect OS
detect_os() {
    if [ -f "/etc/os-release" ]; then
        . /etc/os-release
        echo "$ID"
    elif [ -f "/etc/arch-release" ]; then
        echo "arch"
    elif uname -s | grep -iq "Darwin"; then
        echo "darwin"
    else
        echo "unknown"
    fi
}

OS=$(detect_os)
echo -e "${BLUE}📦 Определена система: ${OS}${NC}"

# Check for WSL
if [ -f "/proc/version" ] && grep -qi "microsoft" /proc/version 2>/dev/null; then
    OS="wsl"
    echo -e "${BLUE}🪟 WSL detected${NC}"
fi

# Install dependencies based on OS
install_deps() {
    case "$OS" in
        ubuntu|debian|raspbian|pop)
            echo -e "${YELLOW}⚙️  Установка зависимостей через apt...${NC}"
            sudo apt update -qq
            sudo apt install -y git curl wget tar gzip golang-go sqlite3 build-essential
            ;;
        fedora|centos|rhel)
            echo -e "${YELLOW}⚙️  Установка зависимостей через dnf/yum...${NC}"
            if command -v dnf &> /dev/null; then
                sudo dnf install -y git curl wget tar gzip golang sqlite-devel gcc gcc-c++ make
            else
                sudo yum install -y git curl wget tar gzip golang sqlite-devel gcc gcc-c++ make
            fi
            ;;
        arch|manjaro|endeavouros)
            echo -e "${YELLOW}⚙️  Установка зависимостей через pacman...${NC}"
            sudo pacman -Syu --noconfirm git curl wget tar gzip go sqlite base-devel
            ;;
        alpine)
            echo -e "${YELLOW}⚙️  Установка зависимостей через apk...${NC}"
            sudo apk add --no-cache git curl wget tar gzip go sqlite-dev build-base
            ;;
        darwin|macos)
            echo -e "${YELLOW}⚙️  Установка зависимостей через Homebrew...${NC}"
            if ! command -v brew &> /dev/null; then
                echo -e "${YELLOW}📦 Установка Homebrew...${NC}"
                /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            fi
            brew install git curl wget go sqlite
            ;;
        wsl)
            echo -e "${YELLOW}⚙️  Установка зависимостей через apt (WSL)...${NC}"
            sudo apt update -qq
            sudo apt install -y git curl wget tar gzip golang-go sqlite3 build-essential
            ;;
        *)
            echo -e "${RED}❌ Неизвестная система: $OS${NC}"
            echo "Пожалуйста, установите вручную: git, curl, wget, go, sqlite3"
            exit 1
            ;;
    esac
}

# Install Go if not available
install_go() {
    if ! command -v go &> /dev/null; then
        echo -e "${YELLOW}📦 Установка Go...${NC}"
        GO_VERSION="1.22.0"
        GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
        GO_URL="https://dl.google.com/go/${GO_TAR}"
        
        # Download Go
        if command -v wget &> /dev/null; then
            wget -q ${GO_URL} -O /tmp/${GO_TAR}
        else
            curl -sL ${GO_URL} -o /tmp/${GO_TAR}
        fi
        
        if [ ! -f /tmp/${GO_TAR} ] || [ ! -s /tmp/${GO_TAR} ]; then
            echo -e "${RED}❌ Не удалось скачать Go${NC}"
            exit 1
        fi
        
        # Extract to /usr/local
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf /tmp/${GO_TAR}
        rm /tmp/${GO_TAR}
        
        # Add to PATH for current session
        export PATH=$PATH:/usr/local/go/bin
        echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile.d/go.sh >/dev/null
    else
        echo -e "${GREEN}✅ Go уже установлен${NC}"
    fi
}

# Install Node.js if not available or < v18
install_node() {
    if ! command -v node &> /dev/null; then
        echo -e "${YELLOW}📦 Установка Node.js 20...${NC}"
        if [ "$OS" == "darwin" ]; then
            brew install node
        else
            curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
            sudo apt install -y nodejs
        fi
    else
        NODE_MAJOR=$(node -v 2>/dev/null | cut -d'v' -f2 | cut -d'.' -f1)
        if [ "$NODE_MAJOR" -lt 18 ]; then
            echo -e "${YELLOW}📦 Обновление Node.js до версии 20...${NC}"
            install_node
        fi
    fi
}

# Install SQLite if not available
install_sqlite() {
    if ! command -v sqlite3 &> /dev/null; then
        echo -e "${YELLOW}📦 Установка SQLite...${NC}"
        if [ "$OS" == "darwin" ]; then
            brew install sqlite
        else
            sudo apt install -y sqlite3 2>/dev/null || \
            sudo dnf install -y sqlite 2>/dev/null || \
            sudo pacman -S --noconfirm sqlite 2>/dev/null
        fi
    fi
}

# Main installation
echo ""
echo -e "${BLUE}🔍 Проверка зависимостей...${NC}"

# Check and install all dependencies
missing_deps=()
for cmd in git curl wget tar gzip; do
    if ! command -v "$cmd" &> /dev/null; then
        missing_deps+=("$cmd")
    fi
done

if [ ${#missing_deps[@]} -gt 0 ]; then
    echo -e "${YELLOW}📦 Отсутствуют системные зависимости: ${missing_deps[*]}${NC}"
    install_deps
else
    echo -e "${GREEN}✅ Все системные зависимости установлены${NC}"
fi

install_go
install_sqlite
install_node

echo ""

# Setup backend
echo -e "${BLUE}⚙️  Настройка backend...${NC}"
cd backend

# Initialize go module if needed
if [ ! -f "go.mod" ]; then
    go mod init github.com/bladekkk/mcap/backend
fi

# Download dependencies
echo -e "${YELLOW}📦 Установка Go зависимостей...${NC}"
go mod tidy

# Verify go.sum exists
if [ ! -f "go.sum" ]; then
    echo -e "${YELLOW}⚠️  Создание go.sum...${NC}"
    go mod tidy
fi

# Run database migrations and start backend
echo -e "${GREEN}🗄️  Миграция базы данных и запуск backend...${NC}"
go run cmd/main.go &
BACKEND_PID=$!
cd ..

# Wait for backend to start
echo -e "${BLUE}📡 Ожидание backend...${NC}"
for i in {1..10}; do
    if curl -s http://localhost:8080/api/v1/servers > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Backend запущен и доступен${NC}"
        break
    fi
    sleep 1
done

# Verify backend
if ! curl -s http://localhost:8080/api/v1/servers > /dev/null 2>&1; then
    echo -e "${RED}❌ Backend не отвечает${NC}"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

# Setup frontend
echo -e "${BLUE}⚙️  Настройка frontend...${NC}"
cd frontend

# Install npm dependencies if not present
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}📦 Установка npm зависимостей (это может занять пару минут)...${NC}"
    npm install
fi

echo -e "${GREEN}🚀 Запуск frontend...${NC}"
npm run dev &
FRONTEND_PID=$!
cd ..

echo ""
echo "========================================"
echo -e "${GREEN}✅ MCAP запущен!${NC}"
echo "========================================"
echo " frontend: ${BLUE}http://localhost:3000${NC}"
echo " backend:  ${BLUE}http://localhost:8080${NC}"
echo "========================================"
echo ""
echo -e "${YELLOW}Нажмите Ctrl+C для остановки всех служб${NC}"
echo ""

# Cleanup on exit
trap "echo; echo -e '${YELLOW}Остановка служб...${NC}'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT TERM

wait