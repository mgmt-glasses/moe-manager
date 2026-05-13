from .models import Character
from .ports import CharacterRepositoryPort

class CharacterUseCase:
    def __init__(self, repository: CharacterRepositoryPort):
        self._repository = repository

    def list_characters(self) -> list[Character]:
        # TODO: implement
        pass

    def get_character(self, character_id: str) -> Character:
        # TODO: implement
        pass
