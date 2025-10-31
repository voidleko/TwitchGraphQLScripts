import requests
import json
import sys
import pprint
from typing import Optional, List
from dataclasses import dataclass
from datetime import datetime
from data import FollowerData


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
          login
          createdAt
        }
      }
    }
  }
}
"""


def get_followers(streamer: str, logging: bool = False) -> Optional[List[FollowerData]]:
	session = requests.Session()
	cursor = None
	result = []

	while True:
		if logging:
			print(f"[LOG:{streamer}] Loaded {len(result)} results")

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
			print(f"[ERROR] Failed request: exit code = {response.status_code}")
			print(response.text)

			return None
		else:
			try:
				data = json.loads(response.text)

				for follower_json in data['data']['user']['followers']['edges']:
					cursor = follower_json['cursor']

					if follower_json['node'] is None:
						# Deleted account
						name = ""
						followed_at = datetime.fromisoformat(follower_json['followedAt'])
						created_at = datetime.now()
						follower = FollowerData(name=name, created_at=created_at, followed_at=followed_at)
						result.append(follower)
					else:
						name = follower_json['node']['login']
						followed_at = datetime.fromisoformat(follower_json['followedAt'])
						created_at = datetime.fromisoformat(follower_json['node']['createdAt'])
						follower = FollowerData(name=name, created_at=created_at, followed_at=followed_at)
						result.append(follower)

				if not data['data']['user']['followers']['pageInfo']['hasNextPage'] or cursor == '':
					break
			except Exception as ex:
				print("[ERROR] Failed: invalid response format for", streamer)
				print("[ERROR]", ex)
				print(response.text)

				return None

	return result


if __name__ == "__main__":
	if len(sys.argv) == 2:
		print(*get_followers(sys.argv[1]), sep="\n")
	else:
		print("Ivalid run format")
