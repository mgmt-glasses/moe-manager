import uuid
from datetime import date, datetime
from typing import Any, Optional

from fastapi import APIRouter, Depends, HTTPException, Query
from pydantic import BaseModel

from moe_screentime.adapters.outbound.repositories.sqlite_screentime_repository import (
    SQLiteScreenTimeRepository,
)
from moe_screentime.domain.models import ScreenTimeRecord
from moe_screentime.domain.use_cases import ScreenTimeUseCase

router = APIRouter(tags=["screentime"])


# ---------------------------------------------------------------------------
# 共通レスポンスヘルパー
# ---------------------------------------------------------------------------

def _success(data: Any) -> dict:
    return {"success": True, "data": data, "error": None}


def _error(code: str, message: str) -> dict:
    return {"success": False, "data": None, "error": {"code": code, "message": message}}


# ---------------------------------------------------------------------------
# 依存性注入
# ---------------------------------------------------------------------------

def get_use_case() -> ScreenTimeUseCase:
    repository = SQLiteScreenTimeRepository()
    return ScreenTimeUseCase(repository=repository)


# ---------------------------------------------------------------------------
# リクエスト / レスポンス DTO
# ---------------------------------------------------------------------------

class EntertainmentRecordRequest(BaseModel):
    minutes: int
    target_minutes: int = 120  # MVP暫定: 本来はユーザ設定から取得


class EntertainmentRecordResponse(BaseModel):
    record_id: str
    user_id: str
    date: date
    minutes: int
    target_minutes: int
    diff_minutes: int


class EntertainmentRecordListItem(BaseModel):
    date: date
    minutes: int
    target_minutes: int
    diff_minutes: int


# ---------------------------------------------------------------------------
# エンドポイント
# ---------------------------------------------------------------------------

@router.put("/users/{user_id}/entertainment-records/{record_date}")
async def upsert_entertainment_record(
    user_id: str,
    record_date: date,
    body: EntertainmentRecordRequest,
    use_case: ScreenTimeUseCase = Depends(get_use_case),
):
    """娯楽時間を登録・更新する (§7.1)"""
    existing = use_case.get_by_date(user_id, record_date)
    record_id = existing.record_id if existing else f"ent_{uuid.uuid4().hex[:8]}"

    record = ScreenTimeRecord(
        record_id=record_id,
        user_id=user_id,
        date=record_date,
        entertainment_minutes=body.minutes,
        target_minutes=body.target_minutes,
    )
    use_case.record(record)

    return _success(
        EntertainmentRecordResponse(
            record_id=record.record_id,
            user_id=record.user_id,
            date=record.date,
            minutes=record.entertainment_minutes,
            target_minutes=record.target_minutes,
            diff_minutes=record.diff_minutes,
        ).model_dump()
    )


@router.get("/users/{user_id}/entertainment-records/{record_date}")
async def get_entertainment_record(
    user_id: str,
    record_date: date,
    use_case: ScreenTimeUseCase = Depends(get_use_case),
):
    """特定日の娯楽時間を取得する (§7.2)"""
    record = use_case.get_by_date(user_id, record_date)
    if record is None:
        raise HTTPException(status_code=404, detail="Record not found")

    return _success(
        EntertainmentRecordResponse(
            record_id=record.record_id,
            user_id=record.user_id,
            date=record.date,
            minutes=record.entertainment_minutes,
            target_minutes=record.target_minutes,
            diff_minutes=record.diff_minutes,
        ).model_dump()
    )


@router.get("/users/{user_id}/entertainment-records")
async def list_entertainment_records(
    user_id: str,
    from_date: Optional[date] = Query(None, alias="from"),
    to_date: Optional[date] = Query(None, alias="to"),
    use_case: ScreenTimeUseCase = Depends(get_use_case),
):
    """期間指定で娯楽時間一覧を取得する (§7.3)"""
    if from_date and to_date:
        records = use_case.get_by_range(user_id, from_date, to_date)
    else:
        records = use_case.get_weekly_summary(user_id)

    items = [
        EntertainmentRecordListItem(
            date=r.date,
            minutes=r.entertainment_minutes,
            target_minutes=r.target_minutes,
            diff_minutes=r.diff_minutes,
        ).model_dump()
        for r in records
    ]
    return _success(items)
