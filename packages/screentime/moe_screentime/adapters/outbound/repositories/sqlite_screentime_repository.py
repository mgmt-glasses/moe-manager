from sqlalchemy import create_engine, Column, String, Integer, Date, desc
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from datetime import date
from typing import Optional

from moe_screentime.domain.models import ScreenTimeRecord

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

    def save(self, record: ScreenTimeRecord) -> None:
        session = self.Session()
        try:
            existing = (
                session.query(ScreenTimeModel)
                .filter(
                    ScreenTimeModel.user_id == record.user_id,
                    ScreenTimeModel.date == record.date,
                )
                .one_or_none()
            )
            if existing is None:
                session.add(
                    ScreenTimeModel(
                        record_id=record.record_id,
                        user_id=record.user_id,
                        date=record.date,
                        entertainment_minutes=record.entertainment_minutes,
                        target_minutes=record.target_minutes,
                    )
                )
            else:
                if record.record_id:
                    existing.record_id = record.record_id
                existing.entertainment_minutes = record.entertainment_minutes
                existing.target_minutes = record.target_minutes
            session.commit()
        finally:
            session.close()

    def get_by_date(self, user_id: str, target_date: date) -> Optional[ScreenTimeRecord]:
        session = self.Session()
        try:
            model = (
                session.query(ScreenTimeModel)
                .filter(
                    ScreenTimeModel.user_id == user_id,
                    ScreenTimeModel.date == target_date,
                )
                .one_or_none()
            )
            return self._to_record(model) if model else None
        finally:
            session.close()

    def list_by_range(self, user_id: str, start_date: date, end_date: date) -> list[ScreenTimeRecord]:
        session = self.Session()
        try:
            models = (
                session.query(ScreenTimeModel)
                .filter(
                    ScreenTimeModel.user_id == user_id,
                    ScreenTimeModel.date >= start_date,
                    ScreenTimeModel.date <= end_date,
                )
                .order_by(ScreenTimeModel.date.asc())
                .all()
            )
            return [self._to_record(model) for model in models]
        finally:
            session.close()

    def list_recent(self, user_id: str, days: int = 7) -> list[ScreenTimeRecord]:
        session = self.Session()
        try:
            models = (
                session.query(ScreenTimeModel)
                .filter(ScreenTimeModel.user_id == user_id)
                .order_by(desc(ScreenTimeModel.date))
                .limit(days)
                .all()
            )
            return [self._to_record(model) for model in models]
        finally:
            session.close()

    def _to_record(self, model: ScreenTimeModel) -> ScreenTimeRecord:
        return ScreenTimeRecord(
            record_id=model.record_id,
            user_id=model.user_id,
            date=model.date,
            entertainment_minutes=model.entertainment_minutes,
            target_minutes=model.target_minutes,
        )
