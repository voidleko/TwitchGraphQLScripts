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
query fetchUser($id: ID, $login: String, $first: Int = 50, $after: Cursor) {
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


def send_request(session: requests.Session, streamer: str, cursor: str):
	MAX_REPEATS = 5
	REPEAT_SLEEP = 1.0

	for i in range(MAX_REPEATS):
		if i > 0:
			time.sleep(REPEAT_SLEEP)
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
			print("[ERROR]", "Status code =", response.status_code)
			continue

		data = json.loads(response.text)
		if 'errors' in data.keys() and len(data['errors']) > 0:
			for error in data['errors']:
				print("[ERROR]", error)
			continue

		return response
	
	return None


def get_followers(streamer: str) -> Optional[List[FollowerData]]:
	session = requests.Session()
	cursor = None
	result = []

	while True:
		print(streamer, ":", len(result))

		response = send_request(session, streamer, cursor)
		if response is None:
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
			print("[ERROR]", ex.__repr__())

			return None

	return result


if __name__ == "__main__":
	if len(sys.argv) == 2:
		print(*get_followers(sys.argv[1]), sep="\n")
	else:
		print("Ivalid run format")
