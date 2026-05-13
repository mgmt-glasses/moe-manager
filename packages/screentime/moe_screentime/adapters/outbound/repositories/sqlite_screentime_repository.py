from sqlalchemy import create_engine, Column, String, Integer, Date
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

Base = declarative_base()

class ScreenTimeModel(Base):
    __tablename__ = "screentime_records"
    record_id = Column(String, primary_key=True)
    user_id = Column(String, nullable=False, index=True)
    date = Column(Date, nullable=False)
    entertainment_minutes = Column(Integer, nullable=False)
    target_minutes = Column(Integer, nullable=False)

class SQLiteScreenTimeRepository:
    def __init__(self, db_path: str = "data/moe.db"):
        self.engine = create_engine(f"sqlite:///{db_path}")
        Base.metadata.create_all(self.engine)
        self.Session = sessionmaker(bind=self.engine)

    # TODO: implement save / get_by_date / list_recent
