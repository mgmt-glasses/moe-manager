from .models import User
from .ports import UserRepositoryPort

class UserUseCase:
    def __init__(self, repository: UserRepositoryPort):
        self._repository = repository

    def register(self, user: User) -> None:
        # TODO: implement
        pass

    def get(self, user_id: str) -> User:
        # TODO: implement
        pass

    def select_character(self, user_id: str, character_id: str) -> None:
        # TODO: implement
        pass
