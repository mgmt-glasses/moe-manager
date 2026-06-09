import uuid
import os
import base64
from datetime import date, datetime
from typing import Any, Optional, List

from fastapi import APIRouter, Depends, HTTPException, Query, Body
from pydantic import BaseModel, Field

from moe_screentime.adapters.outbound.repositories.postgres_screentime_repository import (
    PostgresScreenTimeRepository,
)
from moe_screentime.adapters.outbound.llm.gemini_image_analyzer import GeminiImageAnalyzer
from moe_screentime.domain.models import ScreenTimeRecord, ScreenTimeCategoryUsage
from moe_screentime.domain.use_cases import ScreenTimeUseCase, ScreenTimeImageAnalysisUseCase

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
    repository = PostgresScreenTimeRepository()
    return ScreenTimeUseCase(repository=repository)

def get_image_analysis_use_case() -> ScreenTimeImageAnalysisUseCase:
    api_key = os.environ.get("GEMINI_API_KEY", "")
    analyzer = GeminiImageAnalyzer(api_key=api_key)
    return ScreenTimeImageAnalysisUseCase(analyzer=analyzer)

# ---------------------------------------------------------------------------
# リクエスト / レスポンス DTO
# ---------------------------------------------------------------------------

class CategoryUsageDTO(BaseModel):
    category: str
    minutes: int

class EntertainmentRecordRequest(BaseModel):
    minutes: int
    target_minutes: int = 120  # MVP暫定: 本来はユーザ設定から取得
    categories: List[CategoryUsageDTO] = Field(default_factory=list)


class EntertainmentRecordResponse(BaseModel):
    record_id: str
    user_id: str
    date: date
    minutes: int
    target_minutes: int
    diff_minutes: int
    categories: List[CategoryUsageDTO]


class EntertainmentRecordListItem(BaseModel):
    date: date
    minutes: int
    target_minutes: int
    diff_minutes: int
    categories: List[CategoryUsageDTO]


class ImageAnalysisRequest(BaseModel):
    image_base64: str
    mime_type: str = "image/png"


class ImageAnalysisResponse(BaseModel):
    items: List[CategoryUsageDTO]


# ---------------------------------------------------------------------------
# エンドポイント
# ---------------------------------------------------------------------------

@router.post("/users/{user_id}/entertainment-records/analyze")
async def analyze_entertainment_record_image(
    user_id: str,
    body: ImageAnalysisRequest,
    use_case: ScreenTimeImageAnalysisUseCase = Depends(get_image_analysis_use_case),
):
    """スクリーンタイム画像の解析"""
    try:
        image_bytes = base64.b64decode(body.image_base64)
        result = use_case.analyze(image_bytes, body.mime_type)
        items = [CategoryUsageDTO(category=item.category, minutes=item.minutes) for item in result.items]
        return _success(ImageAnalysisResponse(items=items).model_dump())
    except Exception as e:
        return _error("ANALYSIS_FAILED", str(e))


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

    domain_categories = [
        ScreenTimeCategoryUsage(category=cat.category, minutes=cat.minutes)
        for cat in body.categories
    ]

    record = ScreenTimeRecord(
        record_id=record_id,
        user_id=user_id,
        date=record_date,
        entertainment_minutes=body.minutes,
        target_minutes=body.target_minutes,
        categories=domain_categories,
    )
    use_case.record(record)

    response_categories = [
        CategoryUsageDTO(category=cat.category, minutes=cat.minutes)
        for cat in record.categories
    ]

    return _success(
        EntertainmentRecordResponse(
            record_id=record.record_id,
            user_id=record.user_id,
            date=record.date,
            minutes=record.entertainment_minutes,
            target_minutes=record.target_minutes,
            diff_minutes=record.diff_minutes,
            categories=response_categories,
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

    response_categories = [
        CategoryUsageDTO(category=cat.category, minutes=cat.minutes)
        for cat in record.categories
    ]

    return _success(
        EntertainmentRecordResponse(
            record_id=record.record_id,
            user_id=record.user_id,
            date=record.date,
            minutes=record.entertainment_minutes,
            target_minutes=record.target_minutes,
            diff_minutes=record.diff_minutes,
            categories=response_categories,
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

    items = []
    for r in records:
        cat_dtos = [CategoryUsageDTO(category=cat.category, minutes=cat.minutes) for cat in r.categories]
        items.append(
            EntertainmentRecordListItem(
                date=r.date,
                minutes=r.entertainment_minutes,
                target_minutes=r.target_minutes,
                diff_minutes=r.diff_minutes,
                categories=cat_dtos,
            ).model_dump()
        )

    return _success(items)
