# Scheduled Notification System - Implementation Plan

## Overview

Bot will send Direct Messages (DMs) to configured users about their unread mentions at **11 AM** and **5 PM** daily (Asia/Ho_Chi_Minh timezone).

---

## Requirements

### Core Functionality

**Notification Schedule:**
- **11 AM**: Count unread messages from 5 PM yesterday to 11 AM today
- **5 PM**: Count unread messages from 11 AM today to 5 PM today
- **Timezone**: Asia/Ho_Chi_Minh (UTC+7)

**Unread Message Definition:**
- Message contains `@user`, `@channel`, or `@here` mention
- User has NOT reacted to the message yet
- Tracked per user

**Read Status Detection:**
- Message marked as "read" when user reacts (via `reaction_added` event)
- Update `read_at` timestamp in database
- Remove from Redis unread set

**Re-notification Policy:**
- Allow re-notification if message still unread (even if previously notified)
- Update `notified_at` timestamp after sending each notification

**Target Users:**
- Only users configured in `NOTIFICATION_USERS` env variable
- For `@channel` / `@here`: assume all env users are in all channels
- No API call to fetch channel members

---

## Notification Template

```
You have 10 unread mentions:
  #channel1 (5): View details
  #channel2 (3): View details  
  #channel3 (2): View details
```

**View Details Link:**
- Points to **oldest unread message** in that channel
- Generated using Slack's `chat.getPermalink` API
- Cached in Redis on message arrival

**Example Scenario:**
```
User has 5 unread messages in #general:
- Yesterday 6 PM: Message A
- Yesterday 8 PM: Message B
- Today 9 AM: Message C
- Today 10 AM: Message D
- Today 10:30 AM: Message E

At 11 AM notification → Link points to Message A (oldest)
```

---

## Data Storage Strategy

### Redis (Fast reads for real-time)

**Structure:**
```
Key: "unread:mentions:{user_id}"
Type: Sorted Set
Score: timestamp (float64)
Member: message_id (string)
TTL: 24 hours
```

**Operations:**
- `ZADD` - Add new mention
- `ZCARD` - Count total unread
- `ZRANGE` - Get messages in time range
- `ZREM` - Remove when marked as read

```
Key: "message:{message_id}"
Type: Hash
Fields:
  - channel_id (string)
  - channel_name (string)
  - text_preview (string, max 100 chars)
  - mentioned_users (string, comma-separated)
  - sender_id (string)
  - permalink (string, from chat.getPermalink)
  - timestamp (string)
TTL: 24 hours
```

### MySQL (Persistent record)

**Table: user_mentions**
```sql
CREATE TABLE IF NOT EXISTS user_mentions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    message_id VARCHAR(255) NOT NULL,
    channel_id VARCHAR(255) NOT NULL,
    channel_name VARCHAR(255),
    text_preview TEXT,
    permalink TEXT,
    mentioned_at TIMESTAMP NOT NULL,
    read_at TIMESTAMP NULL,
    notified_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_message_id (message_id),
    INDEX idx_channel_id (channel_id),
    INDEX idx_read_at (read_at),
    INDEX idx_mentioned_at (mentioned_at),
    UNIQUE KEY unique_user_message (user_id, message_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Table: notification_schedules**
```sql
CREATE TABLE IF NOT EXISTS notification_schedules (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    scheduled_time TIMESTAMP NOT NULL,
    notification_count INT DEFAULT 0,
    sent_at TIMESTAMP NULL,
    status ENUM('pending', 'sent', 'failed') DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_scheduled_time (scheduled_time),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

## Implementation Flow

### 1. Message Event Arrives

```go
EventProcessor.handleMessageEvent()
├─> Parse mentions (@user, @channel, @here)
├─> Filter by NOTIFICATION_USERS
├─> Generate permalink (chat.getPermalink API)
├─> Store in Redis
│   ├─> ZADD unread:mentions:{user_id} {timestamp} {message_id}
│   └─> HSET message:{message_id} {fields}
└─> Store in MySQL (user_mentions table)
```

**Mention Parsing Logic:**
```
@user           → Extract user ID (e.g., <@U12345>)
@channel/@here  → Create mention for ALL users in NOTIFICATION_USERS env
```

### 2. Reaction Event Arrives

```go
EventProcessor.handleReactionEvent()
├─> Get reactor user_id and message_id
├─> Update MySQL: SET read_at = NOW() WHERE user_id AND message_id
└─> Update Redis: ZREM unread:mentions:{user_id} {message_id}
```

### 3. Scheduled Notification (11 AM / 5 PM)

```go
NotificationTask.Execute()
├─> For each user in NOTIFICATION_USERS:
│   ├─> Query unread mentions in time window
│   ├─> Group by channel_id
│   ├─> Get oldest message permalink per channel
│   ├─> Build notification message
│   ├─> Send DM via Slack API
│   ├─> Update notified_at in MySQL
│   └─> Log to notification_schedules table
└─> Handle errors with retry logic
```

**Query for Unread Mentions:**
```sql
SELECT * FROM user_mentions
WHERE user_id = ?
AND read_at IS NULL
AND mentioned_at BETWEEN ? AND ?
ORDER BY mentioned_at ASC
```

---

## Scheduled Notification Options

### Using asynq (Recommended)

**Library**: `github.com/hibiken/asynq`

**Benefits:**
- Redis-backed task queue
- Built-in retry logic with exponential backoff
- Cron scheduling support
- Task monitoring and observability
- Distributed worker support

**Setup:**
```go
// Create scheduler
scheduler := asynq.NewScheduler(
    asynq.RedisClientOpt{Addr: redisAddr},
    &asynq.SchedulerOpts{Location: time.FixedZone("Asia/Ho_Chi_Minh", 7*3600)},
)

// Schedule 11 AM daily
scheduler.Register("0 11 * * *", asynq.NewTask("notification:send", nil))

// Schedule 5 PM daily
scheduler.Register("0 17 * * *", asynq.NewTask("notification:send", nil))
```

**Task Handler:**
```go
func HandleNotificationTask(ctx context.Context, t *asynq.Task) error {
    notificationService.SendScheduledNotifications(ctx)
    return nil
}
```

---

## Component Design

### 1. Models

**internal/model/user_mention.go**
```go
type UserMention struct {
    ID            string
    UserID        string
    MessageID     string
    ChannelID     string
    ChannelName   string
    TextPreview   string
    Permalink     string
    MentionedAt   time.Time
    ReadAt        *time.Time
    NotifiedAt    *time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type NotificationSchedule struct {
    ID              string
    UserID          string
    ScheduledTime   time.Time
    NotificationCount int
    SentAt          *time.Time
    Status          string
    ErrorMessage    string
    CreatedAt       time.Time
}
```

### 2. Repositories

**internal/repository/mention_repository.go**
```go
type MentionRepository interface {
    Create(ctx context.Context, mention *model.UserMention) error
    GetUnreadByUser(ctx context.Context, userID string, startTime, endTime time.Time) ([]*model.UserMention, error)
    MarkAsRead(ctx context.Context, userID, messageID string) error
    UpdateNotifiedAt(ctx context.Context, userID string, messageIDs []string) error
}
```

**internal/repository/notification_repository.go**
```go
type NotificationRepository interface {
    CreateSchedule(ctx context.Context, schedule *model.NotificationSchedule) error
    UpdateScheduleStatus(ctx context.Context, id string, status string, sentAt time.Time) error
}
```

### 3. Services

**internal/service/mention_parser.go**
```go
type MentionParser struct {
    notificationUsers []string
}

func (mp *MentionParser) ParseMentions(text string) []string
func (mp *MentionParser) ContainsChannelMention(text string) bool
func (mp *MentionParser) ContainsHereMention(text string) bool
```

**internal/service/notification_service.go**
```go
type NotificationService struct {
    mentionRepo   repository.MentionRepository
    notifRepo     repository.NotificationRepository
    slackClient   *slack.SlackClient
    cache         model.Cache
    logger        *zap.Logger
}

func (ns *NotificationService) SendScheduledNotifications(ctx context.Context) error
func (ns *NotificationService) buildNotificationMessage(mentions []*model.UserMention) string
```

### 4. Slack Integration

**internal/service/slack/slack_client.go**
```go
func (sc *SlackClient) GetPermalink(channelID, messageTS string) (string, error)
func (sc *SlackClient) SendDirectMessage(userID, message string) error
func (sc *SlackClient) GetChannelName(channelID string) (string, error)
```

### 5. Cache Extensions

**pkg/cache/redis.go**
```go
func (r *RedisCache) ZAdd(key string, score float64, member string) error
func (r *RedisCache) ZCard(key string) (int64, error)
func (r *RedisCache) ZRange(key string, start, stop int64) ([]string, error)
func (r *RedisCache) ZRem(key string, members ...string) error
func (r *RedisCache) HSet(key string, values map[string]interface{}) error
func (r *RedisCache) HGetAll(key string) (map[string]string, error)
```

---

## Configuration

### Environment Variables

```bash
# Existing variables
SLACK_BOT_TOKEN=xoxb-...
SLACK_SIGNING_SECRET=...
REDIS_HOST=localhost
REDIS_PORT=6379
MYSQL_DSN=...

# New variables for notifications
NOTIFICATION_USERS=U12345,U67890,U99999
NOTIFICATION_TIMEZONE=Asia/Ho_Chi_Minh
```

### Config Structure

**pkg/config/config.go**
```go
type Config struct {
    // Existing fields...
    
    // Notification settings
    NotificationUsers    []string `mapstructure:"notification_users"`
    NotificationTimezone string   `mapstructure:"notification_timezone"`
}
```

---

## Task Breakdown

### Phase 1: Foundation (High Priority)
- [ ] **Task 1**: Create database migration for user_mentions and notification_schedules tables
- [ ] **Task 2**: Create model structs: UserMention and NotificationSchedule
- [ ] **Task 3**: Add NOTIFICATION_USERS config to env and config package
- [ ] **Task 4**: Extend Redis cache with sorted set operations (ZADD, ZCARD, ZRANGE, ZREM)

### Phase 2: Core Logic (High Priority)
- [ ] **Task 5**: Create MentionRepository for user_mentions CRUD operations
- [ ] **Task 6**: Create NotificationRepository for notification_schedules audit log
- [ ] **Task 7**: Create MentionParser utility to extract @user, @channel, @here from messages
- [ ] **Task 8**: Update EventProcessor.handleMessageEvent to track mentions in Redis + MySQL
- [ ] **Task 9**: Update EventProcessor to cache permalink using chat.getPermalink API
- [ ] **Task 10**: Update EventProcessor.handleReactionEvent to mark messages as read

### Phase 3: Notification Service (High Priority)
- [ ] **Task 11**: Create NotificationService to query unread mentions and build DM message
- [ ] **Task 12**: Implement SlackClient.SendDirectMessage for sending DMs to users
- [ ] **Task 13**: Implement SlackClient.GetPermalink wrapper
- [ ] **Task 14**: Implement SlackClient.GetChannelName helper

### Phase 4: Scheduler (High Priority)
- [ ] **Task 15**: Add asynq dependency to go.mod
- [ ] **Task 16**: Setup asynq server and client for scheduled tasks
- [ ] **Task 17**: Create notification task handler for 11 AM and 5 PM schedules
- [ ] **Task 18**: Wire up all components in cmd/api/main.go

### Phase 5: Testing & Documentation (Medium Priority)
- [ ] **Task 19**: Write unit tests for MentionParser
- [ ] **Task 20**: Write unit tests for NotificationService
- [ ] **Task 21**: Write integration tests for notification flow
- [ ] **Task 22**: Update README.md and .env.example with new configuration

---

## Testing Strategy

### Unit Tests

**Test MentionParser:**
```go
TestParseMentions_WithUserMention()
TestParseMentions_WithChannelMention()
TestParseMentions_WithHereMention()
TestParseMentions_Mixed()
```

**Test NotificationService:**
```go
TestBuildNotificationMessage()
TestSendScheduledNotifications_Success()
TestSendScheduledNotifications_NoUnread()
TestSendScheduledNotifications_Error()
```

### Integration Tests

**Test Full Flow:**
```go
TestNotificationFlow_MessageToNotification()
TestNotificationFlow_ReadViaReaction()
TestNotificationFlow_TimeWindows()
```

---

## Success Criteria

### MVP Success
- ✅ Correctly parses @user, @channel, @here mentions
- ✅ Stores mentions in Redis and MySQL
- ✅ Marks messages as read when user reacts
- ✅ Sends DMs at exactly 11 AM and 5 PM (Asia/Ho_Chi_Minh)
- ✅ Notification message formatted correctly with permalinks
- ✅ Time windows correctly filter unread messages
- ✅ Updates notified_at after sending DMs
- ✅ Only sends to users in NOTIFICATION_USERS env

### Production Ready
- ✅ Handles concurrent message events without race conditions
- ✅ Retry logic for failed notifications (via asynq)
- ✅ Proper error logging and monitoring
- ✅ 80%+ test coverage
- ✅ Documentation complete

---

## Timeline Estimate

**Week 1**: 
- Phase 1 (Foundation): 1-2 days
- Phase 2 (Core Logic): 2-3 days

**Week 2**:
- Phase 3 (Notification Service): 1-2 days
- Phase 4 (Scheduler): 1-2 days
- Phase 5 (Testing & Docs): 2-3 days

**Total**: ~10-14 days for complete implementation

---

## Notes

- Follow Clean Architecture principles
- Use transactions for MySQL operations
- Implement proper error handling with structured logging
- Add correlation IDs for tracing notification flows
- Monitor Redis memory usage (sorted sets + hashes with 24h TTL)
- Consider pagination if user has 100+ unread mentions
- Test timezone handling thoroughly (Asia/Ho_Chi_Minh = UTC+7)
