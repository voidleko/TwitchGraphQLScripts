from dataclasses import dataclass
from datetime import datetime
from typing import Optional, List


@dataclass
class FollowerData:
	name: str
	created_at: datetime
	followed_at: datetime


@dataclass
class FollowData:
    id: int
    login: Optional[str]
    followedAt: datetime


@dataclass
class UserData:
    id: int
    login: Optional[str]
    createdAt: datetime
    deletedAt: Optional[datetime]
    totalCount: int
    follows: List[FollowData]
    