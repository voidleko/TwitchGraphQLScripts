import requests
import json
import sys
from typing import Optional, List
from dataclasses import dataclass
from datetime import datetime
from data import FollowerData


API_URL = "https://gql.twitch.tv/gql"
API_CLIENT_ID = "kd1unb4b3q4t58fwlpcbzcbnm76a8fp"
FOLLOWING_REQUEST_BODY = """
query fetchUser($id: ID, $login: String, $first: Int = 100, $after: Cursor) {
  user(id: $id, login: $login, lookupType: ALL) {
    follows(first: $first, after: $after) {
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


def get_following(user: str) -> Optional[List[FollowerData]]:
	session = requests.Session()
	cursor = None
	result = []
	pagenum = 0

	while True:
		pagenum += 1
		print(f"[LOG] loading following, page {pagenum}")
		
		response = session.post(
			url=API_URL,
			json={
				'query': FOLLOWING_REQUEST_BODY,
				'variables': {
					'login': user,
					'after': cursor
				}
			},
			headers={
				"Client-ID": API_CLIENT_ID
			})

		if response.status_code != 200:
			print(f"Failed request: exit code = {response.status_code}")
			print(response.text)

			return None
		else:
			try:
				data = json.loads(response.text)

				for follower_json in data['data']['user']['follows']['edges']:
					cursor = follower_json['cursor']

					if follower_json['node'] is None:
						# Deleted account
						continue
					else:
						name = follower_json['node']['login']
						followed_at = datetime.fromisoformat(follower_json['followedAt'])
						created_at = datetime.fromisoformat(follower_json['node']['createdAt'])
						follower = FollowerData(name=name, created_at=created_at, followed_at=followed_at)
						result.append(follower)

				if not data['data']['user']['follows']['pageInfo']['hasNextPage']:
					break
			except Exception as e:
				print("Failed: invalid response format")
				print(response.text)

				return None

	return result


if __name__ == "__main__":
	if len(sys.argv) == 2:
		print(*get_following(sys.argv[1]), sep="\n")
	else:
		print("Ivalid run format")
