import requests
import pprint
import json
from datetime import datetime
from typing import Optional, List
from data import UserData, FollowData
from api import API_URL, API_CLIENT_ID

FOLLOWERS_REQUEST_BODY = """
query fetchUser($id: ID, $login: String) {
  user(id: $id, login: $login, lookupType: ALL) {
    id
    login
    createdAt
    deletedAt
    follows(first: 100) {
        totalCount
        edges {
            followedAt
            node {
                id
                login
            }
        }
    }
  }
}
"""


def get_userdata(id: int) -> UserData:
    session = requests.Session()

    response = session.post(
        url=API_URL,
        json={
            'query': FOLLOWERS_REQUEST_BODY,
            'variables': {
                'id': str(id)
            }
        },
        headers={
            "Client-ID": API_CLIENT_ID
        })
    
    assert response.status_code == 200
    
    data = json.loads(response.text)
    if data['data']['user'] is None:
        return None

    id = data['data']['user']['id']
    login = data['data']['user']['login']
    createdAt = data['data']['user']['createdAt']
    deletedAt = data['data']['user']['deletedAt']
    totalCount = data['data']['user']['follows']['totalCount']
    follows = []
    for fdata in data['data']['user']['follows']['edges']:
        followedAt = fdata['followedAt']
        fid = fdata['node']['id']
        flogin = fdata['node']['login']
        follows.append(FollowData(id=int(fid),
                                  login=flogin,
                                  followedAt=datetime.fromisoformat(followedAt)))

    return UserData(id=int(id),
                    login=login,
                    createdAt=datetime.fromisoformat(createdAt),
                    deletedAt=None if deletedAt is None else datetime.fromisoformat(deletedAt),
                    totalCount=int(totalCount),
                    follows=follows)


def get_userdata_by_login(login: str) -> UserData:
    session = requests.Session()

    response = session.post(
        url=API_URL,
        json={
            'query': FOLLOWERS_REQUEST_BODY,
            'variables': {
                'login': login
            }
        },
        headers={
            "Client-ID": API_CLIENT_ID
        })
    
    assert response.status_code == 200
    
    data = json.loads(response.text)
    if data['data']['user'] is None:
        return None

    id = data['data']['user']['id']
    login = data['data']['user']['login']
    createdAt = data['data']['user']['createdAt']
    deletedAt = data['data']['user']['deletedAt']
    totalCount = data['data']['user']['follows']['totalCount']
    follows = []
    for fdata in data['data']['user']['follows']['edges']:
        followedAt = fdata['followedAt']
        fid = fdata['node']['id']
        flogin = fdata['node']['login']
        follows.append(FollowData(id=int(fid),
                                  login=flogin,
                                  followedAt=datetime.fromisoformat(followedAt)))

    return UserData(id=int(id),
                    login=login,
                    createdAt=datetime.fromisoformat(createdAt),
                    deletedAt=None if deletedAt is None else datetime.fromisoformat(deletedAt),
                    totalCount=int(totalCount),
                    follows=follows)
