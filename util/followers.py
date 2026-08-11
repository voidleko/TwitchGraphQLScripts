import requests
import json
import sys
import time
from pprint import pprint
from typing import Optional, List
from data import FollowerData, json_to_follower_data


API_URL = "https://gql.twitch.tv/gql"
API_CLIENT_ID = "kd1unb4b3q4t58fwlpcbzcbnm76a8fp"
FOLLOWERS_REQUEST_BODY = """
query fetchUser($id: ID, $login: String, $first: Int = 100, $after: Cursor) {
  user(id: $id, login: $login, lookupType: ALL) {
    followers(first: $first, after: $after) {
      totalCount
      pageInfo {
        hasNextPage
      }
      edges {
        cursor
        followedAt
        node {
			id
			login
			createdAt
			deletedAt
			follows {
				totalCount
			}
			followers {
				totalCount
			}
        }
      }
    }
  }
}
"""


def send_request(session: requests.Session, 
				 streamer: str, 
				 cursor: str,
				 log: bool,
				 repeat_times: int,
				 repeat_delay: float):
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
					'login': streamer,
					'after': cursor
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


def get_followers(streamer: str,
				  log: bool = False,
				  repeat_times: int = 5,
				  repeat_delay: float = 1.0) -> Optional[List[FollowerData]]:
	session = requests.Session()
	cursor = None
	result = []

	while True:
		print(streamer, ":", len(result))

		response = send_request(session, streamer, cursor, log, repeat_times, repeat_delay)

		if response is None:
			if log:
				print("[ERROR]", "Failed load", streamer)
			return None
		
		try:
			data = json.loads(response.text)
			followers_data = data['data']['user']['followers']
			for follower_data in followers_data['edges']:
				cursor = follower_data['cursor']
				follower = json_to_follower_data(follower_data)
				result.append(follower)

			if not followers_data['pageInfo']['hasNextPage'] or cursor == '':
				break
		except Exception as ex:
			if log:
				print("[ERROR]", ex.__repr__())

			return None

	return result


if __name__ == "__main__":
	if len(sys.argv) == 2:
		print(*get_followers(sys.argv[1]), sep="\n")
	else:
		print("Ivalid run format")
