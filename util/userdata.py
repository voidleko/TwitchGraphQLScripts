import requests
import json
import sys
import time
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


def send_request(session: requests.Session, 
				 login: Optional[str] = None, 
				 id: Optional[int] = None,
				 log: bool = False,
				 repeat_times: int = 5,
				 repeat_delay: float = 1.0):
	for i in range(repeat_times):
		if i > 0:
			time.sleep(repeat_delay)

			if log:
				print("[LOG]", "Repeat last request")

		response = session.post(
			url=API_URL,
			json={
				'query': FOLLOWERS_REQUEST_BODY,
				'variables': {
					'login': login,
					'id': None if id is None else str(id),
				}
			},
			headers={
				"Client-ID": API_CLIENT_ID
			})
		
		if response.status_code != 200:
			if log:
				print("[ERROR]", "Status code =", response.status_code)
			continue

		data = json.loads(response.text)
		if 'errors' in data.keys() and len(data['errors']) > 0:
			for error in data['errors']:
				if log:
					print("[ERROR]", error)
			continue

		return response
	
	return None



def get_userdata(id: int, log: bool = False) -> Optional[UserData]:
    response = send_request(requests.Session(), id=id, log=log)
    assert response.status_code == 200
    
    data = json.loads(response.text)
    if data['data']['user'] is None:
        return None

    user_data = data['data']['user']
    user = json_to_user_data(user_data)

    return user


def get_userdata_by_login(login: str, log: bool = False) -> Optional[UserData]:
    response = send_request(requests.Session(), login=login, log=log)
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
