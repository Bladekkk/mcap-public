import os
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent
DATA_DIR = BASE_DIR.parent / "data"

class Config:
    SECRET_KEY = os.environ.get("SECRET_KEY", "dev-secret-key")
    SQLALCHEMY_DATABASE_URI = os.environ.get("DATABASE_URL", f"sqlite:///{DATA_DIR}/mcap.db")
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    
    RCON_PORT = int(os.environ.get("RCON_PORT", 25575))
    RCON_PASSWORD = os.environ.get("RCON_PASSWORD", "change_me_in_production")
    RCON_TIMEOUT = 5
    
    SERVER_PORT = int(os.environ.get("SERVER_PORT", 8080))
    DATA_DIR = DATA_DIR
