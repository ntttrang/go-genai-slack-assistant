package goldprice

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// MessagePoster is the subset of SlackClient needed to post a message.
type MessagePoster interface {
	PostMessage(channelID, text, threadTS string) (string, string, error)
}

// CronService schedules and sends the daily gold price update to Slack.
type CronService struct {
	scraper   *Scraper
	slack     MessagePoster
	channelID string
	schedule  string
	logger    *zap.Logger
	cron      *cron.Cron
}

// NewCronService creates a CronService. schedule is a standard 5-field cron expression
// (e.g. "0 17 * * *" for 17:00 every day in the server's local timezone).
func NewCronService(scraper *Scraper, slack MessagePoster, channelID, schedule string, logger *zap.Logger) *CronService {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		// Fall back to UTC if the timezone is unavailable.
		loc = time.UTC
	}

	c := cron.New(cron.WithLocation(loc), cron.WithLogger(cron.DiscardLogger))

	return &CronService{
		scraper:   scraper,
		slack:     slack,
		channelID: channelID,
		schedule:  schedule,
		logger:    logger,
		cron:      c,
	}
}

// Start registers the job and begins the scheduler. It returns an error if the
// cron expression is invalid or the channel ID is not configured.
func (cs *CronService) Start() error {
	if cs.channelID == "" {
		cs.logger.Warn("Gold price cron job disabled: GOLD_PRICE_CHANNEL_ID is not set")
		return nil
	}

	_, err := cs.cron.AddFunc(cs.schedule, cs.sendGoldPriceUpdate)
	if err != nil {
		return fmt.Errorf("invalid cron schedule %q: %w", cs.schedule, err)
	}

	cs.cron.Start()
	cs.logger.Info("Gold price cron job started",
		zap.String("schedule", cs.schedule),
		zap.String("channel_id", cs.channelID),
	)
	return nil
}

// Stop gracefully halts the scheduler, waiting for any running job to finish.
func (cs *CronService) Stop() {
	if cs.cron == nil {
		return
	}
	ctx := cs.cron.Stop()
	// Wait up to 30 s for the running job to complete.
	select {
	case <-ctx.Done():
		cs.logger.Info("Gold price cron job stopped")
	case <-time.After(30 * time.Second):
		cs.logger.Warn("Gold price cron job did not stop within 30s")
	}
}

// sendGoldPriceUpdate scrapes prices and posts the formatted message to Slack.
func (cs *CronService) sendGoldPriceUpdate() {
	cs.logger.Info("Running gold price update job")

	data, err := cs.scraper.Scrape()
	if err != nil {
		cs.logger.Error("Failed to scrape gold prices", zap.Error(err))
		// Still attempt to post whatever partial data was collected.
	}

	msg := formatMessage(data)

	_, _, postErr := cs.slack.PostMessage(cs.channelID, msg, "")
	if postErr != nil {
		cs.logger.Error("Failed to post gold price to Slack",
			zap.String("channel_id", cs.channelID),
			zap.Error(postErr),
		)
		return
	}

	cs.logger.Info("Gold price message posted successfully", zap.String("channel_id", cs.channelID))
}

// formatMessage builds the Slack message text from the scraped data.
func formatMessage(data *GoldPriceData) string {
	now := time.Now().In(mustLoadLocation("Asia/Ho_Chi_Minh"))
	dateStr := now.Format("02/01/2006")

	brands := []string{"SJC", "PNJ", "DOJI", "Mi Hồng", "Ngọc Thẩm"}

	var sb strings.Builder
	sb.WriteString(dateStr)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Giá vàng thế giới: %s USD (%s)\n", orNA(data.WorldPrice), orNA(data.WorldChange)))
	sb.WriteString("Giá vàng trong nước: (TPHCM)\n")

	for _, brand := range brands {
		dp, ok := data.Domestic[brand]
		if !ok {
			dp = DomesticPrice{Buy: "N/A", Sell: "N/A"}
		}
		sb.WriteString(fmt.Sprintf("     * %s: Mua vào: %s - Bán ra: %s\n", brand, dp.Buy, dp.Sell))
	}

	return strings.TrimRight(sb.String(), "\n")
}

func orNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}
