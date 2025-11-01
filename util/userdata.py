import requests
import json
import sys
from typing import Optional
from data import UserData, json_to_user_data
from api import API_URL, API_CLIENT_ID

FOLLOWERS_REQUEST_BODY = """
query fetchUser($id: ID, $login: String) {
  user(id: $id, login: $login, lookupType: ALL) {
    id
    login
    createdAt
    deletedAt
    follows {
        totalCount
    }
  }
}
"""


def get_userdata(id: int) -> Optional[UserData]:
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

    user_data = data['data']['user']
    user = json_to_user_data(user_data)

    return user


def get_userdata_by_login(login: str) -> Optional[UserData]:
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
    
    data = json.loads(response.text)
    if data['data']['user'] is None:
        return None

    user_data = data['data']['user']
    user = json_to_user_data(user_data)

    return user


if __name__ == "__main__":
	if len(sys.argv) == 2:
		print(get_userdata_by_login(sys.argv[1]))
	else:
		print("Ivalid run format")
