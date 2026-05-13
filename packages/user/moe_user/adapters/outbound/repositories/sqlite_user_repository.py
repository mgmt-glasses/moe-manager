from sqlalchemy import create_engine, Column, String, Integer, DateTime
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from datetime import datetime

Base = declarative_base()

class UserModel(Base):
    __tablename__ = "users"
    user_id = Column(String, primary_key=True)
    username = Column(String, nullable=False)
    boss_name = Column(String, nullable=False)
    selected_character_id = Column(String, nullable=True)
    target_entertainment_minutes = Column(Integer, default=120)
    created_at = Column(DateTime, default=datetime.utcnow)

class SQLiteUserRepository:
    def __init__(self, db_path: str = "data/moe.db"):
        self.engine = create_engine(f"sqlite:///{db_path}")
        Base.metadata.create_all(self.engine)
        self.Session = sessionmaker(bind=self.engine)

    # TODO: implement save / get / update_selected_character
