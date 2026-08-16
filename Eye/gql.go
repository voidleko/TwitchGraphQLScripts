package main

const (
	VtubersPageQuery = `
query fetchVtubers($first: Int, $options: StreamOptions, $after: Cursor) {
  streams(first: $first, options: $options, after: $after) {
    edges {
      cursor
      node {
        id
        viewersCount
        previewImageURL(width: 640, height: 360)
        broadcaster {
          id
          login
          broadcastSettings {
            title
          }
        }
      }
    }
    pageInfo {
      hasNextPage
    }
  }
}
`
	UserQuery = `
query fetchUser($id: ID, $login: String) {
  user(id: $id, login: $login, lookupType: ALL) {
    id
    login
  }
}
`
	ViewersPageQuery = `
query fetchViewers($id: ID, $login: String) {
  user(id: $id, login: $login, lookupType: ALL) {
    id
    login
    stream {
      createdAt
      viewersCount
    }
    broadcastSettings {
      title
    }
    channel {
      chatters {
        count
        viewers {
          login
        }
      }
    }
  }
}
`
)
