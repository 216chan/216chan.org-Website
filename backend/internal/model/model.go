package model

import "time"

type Stats struct {
	TotalPosts    int64  `json:"total_posts"`
	CurrentUsers  int64  `json:"current_users"`
	ActiveContent string `json:"active_content"`
}

type Board struct {
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsNSFW      bool      `json:"is_nsfw"`
	CreatedAt   time.Time `json:"created_at"`
}

type Thread struct {
	ID         int64     `json:"id"`
	BoardSlug  string    `json:"board_slug"`
	IsPinned   bool      `json:"is_pinned"`
	IsLocked   bool      `json:"is_locked"`
	ReplyCount int       `json:"reply_count"`
	BumpedAt   time.Time `json:"bumped_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type Post struct {
	ID           int64     `json:"id"`
	ThreadID     int64     `json:"thread_id"`
	BoardSlug    string    `json:"board_slug"`
	IsOP         bool      `json:"is_op"`
	Body         string    `json:"body"`
	ImageURL     *string   `json:"image_url"`
	ImageName    *string   `json:"image_name"`
	FileSize     *int      `json:"file_size_bytes"`
	CountryCode  string    `json:"country_code"`
	CreatedAt    time.Time `json:"created_at"`
}

type ThreadWithOP struct {
	Thread Thread `json:"thread"`
	OP     *Post  `json:"op"`
	Replies []Post `json:"replies,omitempty"`
}

type ThreadsPage struct {
	Board    Board          `json:"board"`
	Threads  []ThreadWithOP `json:"threads"`
	Page     int            `json:"page"`
	Limit    int            `json:"limit"`
	Total    int            `json:"total"`
	Pages    int            `json:"pages"`
}

type ThreadPage struct {
	Board   Board  `json:"board"`
	Thread  Thread `json:"thread"`
	OP      Post   `json:"op"`
	Replies []Post `json:"replies"`
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Total   int    `json:"total"`
	Pages   int    `json:"pages"`
}