import json
from google import genai
from google.genai import types

from moe_screentime.domain.models import ScreenTimeAnalysisResult, ScreenTimeCategoryUsage
from moe_screentime.domain.ports import ScreenTimeImageAnalyzerPort

class GeminiImageAnalyzer(ScreenTimeImageAnalyzerPort):
    def __init__(self, api_key: str):
        self.client = genai.Client(api_key=api_key)

    def analyze(self, image_bytes: bytes, mime_type: str = "image/png") -> ScreenTimeAnalysisResult:
        prompt = """
        これはスマートフォンのスクリーンタイム（利用時間）のスクリーンショットです。
        アプリごとの利用時間を読み取り、以下のジャンルに分類して、それぞれの合計利用時間（分）を計算してください。

        【分類するジャンル（iPhone標準準拠）】
        - エンターテイメント
        - ソーシャル
        - ゲーム
        - 仕事効率化
        - クリエイティビティ
        - 教育
        - ユーティリティ
        - その他

        出力は必ず以下の形式のJSONのみとしてください。マークダウンのバッククォート(```json ... ```)を含めないでください。
        [
            {"category": "ジャンル名", "minutes": 分数},
            ...
        ]
        """

        response = self.client.models.generate_content(
            model='gemini-2.5-flash',
            contents=[
                types.Part.from_bytes(data=image_bytes, mime_type=mime_type),
                prompt,
            ],
            config=types.GenerateContentConfig(
                response_mime_type="application/json",
            )
        )
        
        try:
            # GeminiはJSON形式で返してくるのでパースする
            data = json.loads(response.text)
            items = []
            for item in data:
                items.append(ScreenTimeCategoryUsage(
                    category=item.get("category", "不明"),
                    minutes=int(item.get("minutes", 0))
                ))
            return ScreenTimeAnalysisResult(items=items)
        except Exception as e:
            raise RuntimeError(f"Failed to parse Gemini response: {e}\nResponse: {response.text}")
