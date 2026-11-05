# MCAP - Minecraft Admin Panel

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python 3.12](https://img.shields.io/badge/Python-3.12-blue)](https://python.org/)
[![Node.js 18+](https://img.shields.io/badge/Node.js-18+-green)](https://nodejs.org/)
[![Flask 3.0](https://img.shields.io/badge/Flask-3.0-red)](https://flask.palletsprojects.com/)
[![Next.js 16](https://img.shields.io/badge/Next.js-16-black)](https://nextjs.org/)

Продвинутая панель управления Minecraft серверами с поддержкой множества типов серверов (Vanilla, Paper, Fabric, Forge, Quilt, Purpur, Velocity, Waterfall) и автоматическим управлением зависимостями.

## 🚀 Возможности

- **🎯 Управление серверами**: Создание, запуск, остановка, перезапуск серверов с одного клика
- **📊 Мониторинг в реальном времени**: Live TPS, RAM, CPU, онлайн игроки через WebSocket
- **🔌 Управление плагинами**: Поиск и установка плагинов из Modrinth, Hangar, CurseForge
- **☕ Автоматическое управление Java**: Автоматический выбор правильной версии Java для типа сервера
- **🔐 Безопасность**: JWT аутентификация, защита RCON паролем, логирование действий
- **⚡ Real-time Dashboard**: Современный интерфейс с Tailwind CSS и React

## 🏗️ Технический стек

### Backend
- **Python 3.12** с Flask 3.0
- **SQLite** база данных с миграциями
- **WebSocket** для real-time обновлений
- **Native subprocess** управление Minecraft серверами

### Frontend
- **Next.js 16** с App Router и React 19
- **TypeScript** для type-safety
- **Tailwind CSS 4** + Turbopack
- **Real-time fetching** для live обновлений

## 📋 Структура проекта

```
MCAP-project/
├── backend/                    # Flask приложение
│   ├── cmd/                   # Entry point (планируется)
│   ├── services/              # Бизнес логика
│   │   ├── server_manager.py
│   │   ├── java_manager.py
│   │   ├── downloader.py
│   │   └── plugin_aggregator.py
│   ├── app.py                 # Flask приложение
│   ├── models.py              # SQLAlchemy моделі
│   ├── routes.py              # API endpoints
│   ├── config.py              # Конфиг
│   ├── database_init.py       # DB инициализация
│   ├── requirements.txt
│   └── venv/                  # Virtual environment
│
├── frontend/                   # Next.js приложение
│   ├── app/
│   │   ├── page.tsx           # Главная страница
│   │   ├── layout.tsx         # Layout
│   │   └── globals.css        # Глобальные стили
│   ├── public/
│   ├── package.json
│   └── tsconfig.json
│
├── data/                       # Данные приложения
│   ├── mcap.db               # SQLite база
│   ├── cache/                # Кэш для загрузок
│   ├── jdk/                  # Java установки
│   └── servers/              # Данные серверов
│
├── start.sh                    # Запуск всех сервисов
├── stop.sh                     # Остановка всех сервисов
└── README.md
```

## 📦 Системные требования

| ОС | Версия | Статус |
|----|---------| -------|
| Ubuntu | 20.04+ | ✅ Поддерживается |
| Debian | 11+ | ✅ Поддерживается |
| Fedora | 36+ | ✅ Поддерживается |
| Arch | rolling | ✅ Поддерживается |
| macOS | 12+ | ✅ Поддерживается |
| WSL | Windows 10/11 | ✅ Поддерживается |

## 🚀 Быстрый старт

### Необходимо

- Python 3.12+
- Node.js 18+
- pip и npm

### Установка и запуск

```bash
# Клонировать репозиторий
git clone https://github.com/Bladekkk/mcap-public.git
cd MCAP-project

# Установить зависимости backend
cd backend
python3 -m venv venv
source venv/bin/activate  # На Windows: venv\Scripts\activate
pip install -r requirements.txt

# Установить зависимости frontend (новый терминал)
cd ../frontend
npm install

# Запустить оба сервиса одной командой
cd ..
chmod +x start.sh
./start.sh
```

Или запустить вручную:

```bash
# Terminal 1 - Backend
cd backend
source venv/bin/activate
python3 app.py

# Terminal 2 - Frontend
cd frontend
npm run dev
```

## 🌐 Доступ

- **Dashboard**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/v1
- **WebSocket**: ws://localhost:8080/ws

## 📡 API Endpoints

### Серверы
| Метод | Endpoint | Описание |
|-------|----------|---------|
| GET | `/api/v1/servers` | Список всех серверов |
| POST | `/api/v1/servers` | Создать новый сервер |
| GET | `/api/v1/servers/:id` | Детали сервера |
| POST | `/api/v1/servers/:id/start` | Запустить сервер |
| POST | `/api/v1/servers/:id/stop` | Остановить сервер |
| POST | `/api/v1/servers/:id/restart` | Перезапустить сервер |
| DELETE | `/api/v1/servers/:id` | Удалить сервер |

### Java
| Метод | Endpoint | Описание |
|-------|----------|---------|
| GET | `/api/v1/java` | Список установленных JDK |
| POST | `/api/v1/java` | Установить JDK |

### Метрики
| Метод | Endpoint | Описание |
|-------|----------|---------|
| GET | `/api/v1/metrics/:id` | Real-time метрики |
| GET | `/api/v1/metrics/:id/history` | История метрик |

### Плагины
| Метод | Endpoint | Описание |
|-------|----------|---------|
| GET | `/api/v1/plugins?q=...` | Поиск плагинов |
| POST | `/api/v1/plugins/download` | Загрузить плагин |

### Загрузки
| Метод | Endpoint | Описание |
|-------|----------|---------|
| POST | `/api/v1/downloads` | Загрузить JAR |

### WebSocket
| Event | Описание |
|-------|---------|
| `connect` | Клиент подключен |
| `disconnect` | Клиент отключен |
| `join` | Присоединиться к комнате сервера |
| `leave` | Покинуть комнату сервера |

## ⚙️ Конфигурация

Создайте файл `.env` в `backend/` директории:

```env
SECRET_KEY=your-secret-key-here
DATABASE_URL=sqlite:///../data/mcap.db
SQLALCHEMY_TRACK_MODIFICATIONS=false

SERVER_PORT=8080
RCON_PORT=25575
RCON_PASSWORD=your-rcon-password
RCON_TIMEOUT=5

MODRINTH_TOKEN=your_token_here
CURSEFORGE_KEY=your_key_here
```

## 📊 Примеры использования

### Создание сервера

```bash
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Survival Server",
    "jar_path": "/path/to/server.jar",
    "port": 25565,
    "java_version": "17",
    "xms_mb": 1024,
    "xmx_mb": 4096,
    "eula_accepted": true
  }'
```

### Запуск сервера

```bash
curl -X POST http://localhost:8080/api/v1/servers/{server_id}/start
```

### Получение метрик

```bash
curl http://localhost:8080/api/v1/metrics/{server_id}
```

### Поиск плагинов

```bash
curl "http://localhost:8080/api/v1/plugins?q=EssentialsX&loaders=paper,spigot"
```

## 🐛 Известные ограничения

1. **Rate limiting**: Modrinth/CurseForge имеют ограничения на количество запросов
2. **Cloudflare блокировка**: SpigotMC может блокировать автоматические запросы
3. **Platform зависимость**: `taskset` работает только на Linux
4. **RCON требование**: Нужно `enable-rcon=true` в server.properties

## 🔄 Стан разработки

- [x] Базовое управление серверами
- [x] API endpoints
- [x] Frontend dashboard
- [x] Database инициализация
- [ ] Аутентификация и JWT
- [ ] WebSocket console streaming
- [ ] UI для установки плагинов
- [ ] Автоматические бэкапы
- [ ] Логирование событий
- [ ] Metrics графики
- [ ] Docker поддержка
- [ ] Кластеризация

## 📝 Логирование

Логи находятся в корневой директории проекта:

```bash
# Backend логи
tail -f backend.log

# Frontend логи  
tail -f frontend.log
```

## 🛠️ Разработка

### Установка dev зависимостей

```bash
# Backend
cd backend
pip install -r requirements.txt
pip install pytest pytest-cov  # Для тестів

# Frontend
cd frontend
npm install --save-dev @testing-library/react jest
```

### Запуск тестов

```bash
# Backend
cd backend
pytest

# Frontend
cd frontend
npm run test
```

### Сборка для продакшена

```bash
# Backend
cd backend
source venv/bin/activate
python3 app.py  # Вже готов до продакшена

# Frontend
cd frontend
npm run build
npm start
```

## 🐳 Docker (планируется)

```bash
docker-compose up --build
```

## 📄 Лицензия

MIT License - см. [LICENSE](LICENSE)

## 👨‍💻 Автор

- **Bladekkk** - Initial work

## 🤝 Контрибьютинг

Приветствуются pull requests. Для больших изменений сначала откройте issue для обсуждения.

## 📞 Поддержка

Если у вас есть вопросы или проблемы:

1. Проверьте [STATUS.md](STATUS.md) для информации о текущем состоянии
2. Откройте [GitHub Issue](https://github.com/Bladekkk/mcap-public/issues)
3. Проверьте логи: `backend.log` и `frontend.log`

## 🔐 Безопасность

- ❗ Измените `SECRET_KEY` в продакшене
- ❗ Используйте сильный `RCON_PASSWORD`
- ❗ Не публикуйте `.env` файлы
- ✅ Всегда запускайте обновления перед использованием

## 📈 Производительность

- Backend использует асинхронные операции через Eventlet
- Frontend оптимизирован с помощью Turbopack и Next.js 16
- База данных кэшируется для быстрого доступа
- Real-time обновления через WebSocket

## 🎯 Roadmap

- Q1 2026: WebSocket консоль
- Q2 2026: Аутентификация и RBAC
- Q3 2026: Docker поддержка
- Q4 2026: Clustering

---

**Версія**: 0.1.0  
**Статус**: ✅ Beta  
**Останнє оновлення**: 2026-11-05
