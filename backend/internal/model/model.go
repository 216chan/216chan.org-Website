package model

import "time"

type Stats struct {
	VisibleThreads int64 `json:"visible_threads"`
	TotalPosts     int64 `json:"total_posts"`
	PostsLastDay   int64 `json:"posts_24h"`
	MediaBytes     int64 `json:"visible_media_bytes"`
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
	Permalink   string    `json:"permalink,omitempty"`
	ID          int64     `json:"id"`
	ThreadID    int64     `json:"thread_id"`
	BoardSlug   string    `json:"board_slug"`
	IsOP        bool      `json:"is_op"`
	Body        string    `json:"body"`
	Media       *Media    `json:"media"`
	CountryCode string    `json:"country_code"`
	CreatedAt   time.Time `json:"created_at"`
}

type Media struct {
	URL    string `json:"url"`
	Name   string `json:"name"`
	Size   int    `json:"size_bytes"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type ThreadWithOP struct {
	Thread  Thread `json:"thread"`
	OP      Post   `json:"op"`
	Replies []Post `json:"replies"`
}

type Pagination struct {
	Page  int `json:"page"`
	Pages int `json:"pages"`
	Total int `json:"total"`
}

type ThreadsPage struct {
	Board   Board          `json:"board"`
	Threads []ThreadWithOP `json:"threads"`
	Pagination
}

type ThreadPage struct {
	Board   Board  `json:"board"`
	Thread  Thread `json:"thread"`
	OP      Post   `json:"op"`
	Replies []Post `json:"replies"`
	Pagination
}
