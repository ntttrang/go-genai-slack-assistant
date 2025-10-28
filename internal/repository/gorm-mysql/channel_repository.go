package gormmysql

import (
	"database/sql"
	"fmt"

	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/ntttrang/go-genai-slack-assistant/internal/service"
)

// ChannelRepositoryImpl implements service.ChannelRepository interface
type ChannelRepositoryImpl struct {
	db *sql.DB
}

// NewChannelRepository creates a new channel repository instance
func NewChannelRepository(db *sql.DB) service.ChannelRepository {
	return &ChannelRepositoryImpl{db: db}
}

func (cr *ChannelRepositoryImpl) Save(config *model.ChannelConfig) error {
	languages := ""
	for i, lang := range config.SourceLanguages {
		if i > 0 {
			languages += ","
		}
		languages += lang
	}

	query := `
		INSERT INTO channel_configs (channel_id, auto_translate, source_languages, target_language, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`

	result, err := cr.db.Exec(query, config.ChannelID, config.AutoTranslate, languages, config.TargetLanguage, config.Enabled)
	if err != nil {
		return fmt.Errorf("failed to save channel config: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	config.ID = fmt.Sprintf("%d", id)
	return nil
}

func (cr *ChannelRepositoryImpl) GetByChannelID(channelID string) (*model.ChannelConfig, error) {
	query := `
		SELECT id, channel_id, auto_translate, source_languages, target_language, enabled, created_at, updated_at
		FROM channel_configs
		WHERE channel_id = ?
	`

	config := &model.ChannelConfig{}
	var languages string

	err := cr.db.QueryRow(query, channelID).Scan(
		&config.ID,
		&config.ChannelID,
		&config.AutoTranslate,
		&languages,
		&config.TargetLanguage,
		&config.Enabled,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("channel config not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channel config: %w", err)
	}

	// Parse comma-separated languages
	if languages != "" {
		config.SourceLanguages = parseLanguages(languages)
	}

	return config, nil
}

func (cr *ChannelRepositoryImpl) Update(config *model.ChannelConfig) error {
	languages := ""
	for i, lang := range config.SourceLanguages {
		if i > 0 {
			languages += ","
		}
		languages += lang
	}

	query := `
		UPDATE channel_configs
		SET auto_translate = ?, source_languages = ?, target_language = ?, enabled = ?, updated_at = NOW()
		WHERE channel_id = ?
	`

	result, err := cr.db.Exec(query, config.AutoTranslate, languages, config.TargetLanguage, config.Enabled, config.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to update channel config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("channel config not found")
	}

	return nil
}

func (cr *ChannelRepositoryImpl) Delete(channelID string) error {
	query := `DELETE FROM channel_configs WHERE channel_id = ?`

	result, err := cr.db.Exec(query, channelID)
	if err != nil {
		return fmt.Errorf("failed to delete channel config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("channel config not found")
	}

	return nil
}

func (cr *ChannelRepositoryImpl) GetAll() ([]*model.ChannelConfig, error) {
	query := `
		SELECT id, channel_id, auto_translate, source_languages, target_language, enabled, created_at, updated_at
		FROM channel_configs
		ORDER BY created_at DESC
	`

	rows, err := cr.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query channel configs: %w", err)
	}
	defer rows.Close()

	var configs []*model.ChannelConfig

	for rows.Next() {
		config := &model.ChannelConfig{}
		var languages string

		err := rows.Scan(
			&config.ID,
			&config.ChannelID,
			&config.AutoTranslate,
			&languages,
			&config.TargetLanguage,
			&config.Enabled,
			&config.CreatedAt,
			&config.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan channel config: %w", err)
		}

		if languages != "" {
			config.SourceLanguages = parseLanguages(languages)
		}

		configs = append(configs, config)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return configs, nil
}

func parseLanguages(str string) []string {
	var langs []string
	for i := 0; i < len(str); i++ {
		if str[i] == ',' {
			continue
		}
		j := i
		for j < len(str) && str[j] != ',' {
			j++
		}
		langs = append(langs, str[i:j])
		i = j
	}
	return langs
}
