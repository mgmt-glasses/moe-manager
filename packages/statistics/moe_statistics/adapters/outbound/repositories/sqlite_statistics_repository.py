from datetime import date
# statistics モジュールは task / screentime の DB を横断集計するため、
# 独自テーブルは持たず moe_task / moe_screentime のDBを参照するクエリアダプタとして実装する。
# （同一 SQLite ファイルを参照することで JOIN が可能）

class SQLiteStatisticsRepository:
    def __init__(self, db_path: str = "data/moe.db"):
        self.db_path = db_path

    # TODO: implement get_daily_summary / get_weekly_summaries
    #   - tasks テーブルを集計して完了数・未完了数・完了率を算出
    #   - screentime_records テーブルから当日の娯楽時間を取得
