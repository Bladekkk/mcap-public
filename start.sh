#!/bin/bash
set -e

echo "========================================"
echo "MCAP - Minecraft Server Management Panel"
echo "========================================"
echo ""

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Проверка зависимостей
check_dependency() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}❌ $1 не найден. Установите его и попробуйте снова.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ $1${NC}"
}

echo -e "${BLUE}🔍 Проверка зависимостей...${NC}"
check_dependency "go"
check_dependency "node"
check_dependency "npm"
check_dependency "sqlite3"
echo ""

# Настройка backend
echo -e "${BLUE}⚙️  Настройка backend...${NC}"
cd backend

# Если нет go.sum, скачиваем зависимости
if [ ! -f "go.sum" ]; then
    echo -e "${YELLOW}📦 Установка зависимостей Go...${NC}"
    go mod tidy
fi

echo -e "${GREEN}🚀 Запуск backend...${NC}"
go run cmd/main.go &
BACKEND_PID=$!
sleep 3
cd ..

# Проверка доступности backend
echo -e "${BLUE}📡 Проверка backend...${NC}"
if curl -s http://localhost:8080/api/v1/servers > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Backend доступен на http://localhost:8080${NC}"
else
    echo -e "${RED}❌ Backend не отвечает${NC}"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

# Настройка frontend
echo -e "${BLUE}⚙️  Настройка frontend...${NC}"
cd frontend

# Установка npm зависимостей
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}📦 Установка зависимостей Node.js...${NC}"
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

# Обработка завершения
trap "echo; echo -e '${YELLOW}Остановка служб...${NC}'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT TERM

wait
