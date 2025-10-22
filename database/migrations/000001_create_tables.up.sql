CREATE TABLE IF NOT EXISTS channel_configs (
    id VARCHAR(36) PRIMARY KEY,
    channel_id VARCHAR(255) NOT NULL UNIQUE,
    auto_translate BOOLEAN DEFAULT FALSE,
    source_languages JSON,
    target_language VARCHAR(10) DEFAULT 'vi',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_channel_id (channel_id),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS translations (
    id VARCHAR(36) PRIMARY KEY,
    source_message_id VARCHAR(255) NOT NULL,
    source_text LONGTEXT NOT NULL,
    source_language VARCHAR(10),
    target_language VARCHAR(10),
    translated_text LONGTEXT NOT NULL,
    hash VARCHAR(64) NOT NULL,
    user_id VARCHAR(255),
    channel_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ttl BIGINT,
    INDEX idx_hash (hash),
    INDEX idx_channel_id (channel_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS usage_metrics (
    id VARCHAR(36) PRIMARY KEY,
    channel_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    translation_count INT DEFAULT 0,
    api_calls INT DEFAULT 0,
    cache_hits INT DEFAULT 0,
    cache_misses INT DEFAULT 0,
    error_count INT DEFAULT 0,
    total_latency_ms BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_channel_id (channel_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
