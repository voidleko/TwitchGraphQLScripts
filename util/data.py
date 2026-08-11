from dataclasses import dataclass
from datetime import datetime
from typing import Optional, List


@dataclass
class UserData:
    id: int
    login: Optional[str]
    created_at: datetime
    deleted_at: Optional[datetime]
    follows_count: int
    

@dataclass
class FollowerData:
    user: Optional[UserData]  # None if user data is empty, when user banned or deleted
    followed_at: datetime


def json_to_user_data(data):
    id = data['id']
    login = data['login']
    created_at = data['createdAt']
    deleted_at = data['deletedAt']
    follows_count = data['follows']['totalCount']

    id = int(id)
    created_at = datetime.fromisoformat(created_at)
    if deleted_at is not None:
        deleted_at = datetime.fromisoformat(deleted_at)

    return UserData(id=id,
                    login=login,
                    created_at=created_at,
                    deleted_at=deleted_at,
                    follows_count=follows_count)


def json_to_follower_data(data):
    followed_at = datetime.fromisoformat(data['followedAt'])

    if data['node'] is None:
        return FollowerData(user=None,
                            followed_at=followed_at)
    else:
        user_data = json_to_user_data(data['node'])

        return FollowerData(user=user_data,
                            followed_at=followed_at)
