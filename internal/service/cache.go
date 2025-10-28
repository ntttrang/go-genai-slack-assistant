package service

// Cache defines the interface for cache operations
type Cache interface {
	Get(key string) (string, error)
	Set(key string, value string, ttl int64) error
	Delete(key string) error
	Exists(key string) (bool, error)
}
