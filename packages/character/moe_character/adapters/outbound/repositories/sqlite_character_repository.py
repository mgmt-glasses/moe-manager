from sqlalchemy import create_engine, Column, String
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

Base = declarative_base()

class CharacterModel(Base):
    __tablename__ = "characters"
    character_id = Column(String, primary_key=True)
    name = Column(String, nullable=False)
    mbti = Column(String, nullable=False)
    description = Column(String)
    speech_style = Column(String)
    voice_preset_id = Column(String)
    icon_path = Column(String, nullable=True)
    voice_sample_path = Column(String, nullable=True)

class SQLiteCharacterRepository:
    def __init__(self, db_path: str = "data/moe.db"):
        self.engine = create_engine(f"sqlite:///{db_path}")
        Base.metadata.create_all(self.engine)
        self.Session = sessionmaker(bind=self.engine)

    # TODO: implement get_all / get / get_by_mbti
