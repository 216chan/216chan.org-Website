package model

type Stats struct {
	TotalPosts    int64  `json:"total_posts"`
	CurrentUsers  int64  `json:"current_users"`
	ActiveContent string `json:"active_content"`
}
