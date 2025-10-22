package model

type RateLimiter interface {
	CheckUserLimit(userID string) (allowed bool, remaining int, resetTime int64, err error)
	CheckChannelLimit(channelID string) (allowed bool, remaining int, resetTime int64, err error)
	IncrementUserLimit(userID string) error
	IncrementChannelLimit(channelID string) error
}
