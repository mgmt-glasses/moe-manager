from .models import Task
from .ports import TaskRepositoryPort

class TaskUseCase:
    def __init__(self, repository: TaskRepositoryPort):
        self._repository = repository

    def create(self, task: Task) -> None:
        # TODO: implement
        pass

    def complete(self, task_id: str) -> None:
        # TODO: implement
        pass

    def delete(self, task_id: str) -> None:
        # TODO: implement
        pass

    def list_today(self, user_id: str) -> list[Task]:
        # TODO: implement
        pass

    def get_today_summary(self, user_id: str) -> dict:
        # Returns {"total": int, "completed": int, "completion_rate": float}
        # TODO: implement
        pass
