from flask import Flask
from flask_cors import CORS
from database_init import db, socketio

def create_app():
    app = Flask(__name__)
    app.config.from_object("config.Config")
    
    CORS(app, resources={r"/api/*": {"origins": "http://localhost:3000"}})
    
    db.init_app(app)
    socketio.init_app(app, cors_allowed_origins="*")
    
    from routes import api_bp
    app.register_blueprint(api_bp, url_prefix="/api/v1")
    
    with app.app_context():
        db.create_all()
    
    return app

app = create_app()

if __name__ == "__main__":
    socketio.run(app, host="0.0.0.0", port=8080)
