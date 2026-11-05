#!/bin/bash
set -e

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

# Install system dependencies
echo -e "${BLUE}🔍 Установка системных зависимостей...${NC}"
case "$OS" in
    ubuntu|debian|raspbian|pop)
        sudo apt update -qq
        sudo apt install -y python3 python3-pip python3-venv git curl wget tar gzip sqlite3
        ;;
    fedora|centos|rhel)
        if command -v dnf &> /dev/null; then
            sudo dnf install -y python3 python3-pip git curl wget tar gzip sqlite
        else
            sudo yum install -y python3 python3-pip git curl wget tar gzip sqlite
        fi
        ;;
    arch|manjaro|endeavouros)
        sudo pacman -Syu --noconfirm python python-pip git curl wget tar gzip sqlite
        ;;
    darwin|macos)
        if ! command -v brew &> /dev/null; then
            echo -e "${YELLOW}📦 Установка Homebrew...${NC}"
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        fi
        brew install python git curl wget sqlite
        ;;
    *)
        echo -e "${RED}❌ Неизвестная система: $OS${NC}"
        exit 1
        ;;
esac

echo -e "${GREEN}✅ Системные зависимости установлены${NC}"
echo ""

# Setup backend
echo -e "${BLUE}⚙️  Настройка backend...${NC}"
cd backend

# Create virtual environment
echo -e "${YELLOW}📦 Создание виртуального окружения Python...${NC}"
python3 -m venv venv
source venv/bin/activate
pip install --upgrade pip

# Install Python dependencies
echo -e "${YELLOW}📦 Установка Python зависимостей...${NC}"
pip install -r requirements.txt

# Create data directories
mkdir -p data/servers data/cache

# Run server
echo -e "${GREEN}🚀 Запуск backend...${NC}"
FLASK_APP=app.py flask run --host=0.0.0.0 --port=8080 &
BACKEND_PID=$!
cd ..

# Wait for backend
echo -e "${BLUE}📡 Ожидание backend...${NC}"
sleep 3

# Verify backend
if ! curl -s http://localhost:8080/api/v1/servers > /dev/null 2>&1; then
    echo -e "${RED}❌ Backend не отвечает${NC}"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

echo -e "${GREEN}✅ Backend доступен${NC}"
echo ""

# Setup frontend
echo -e "${BLUE}⚙️  Настройка frontend...${NC}"
cd frontend

if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}📦 Установка npm зависимостей...${NC}"
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

trap "echo; echo -e '${YELLOW}Остановка служб...${NC}'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT TERM

wait
