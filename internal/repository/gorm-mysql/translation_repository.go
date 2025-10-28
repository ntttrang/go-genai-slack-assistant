package gormmysql

import (
	"database/sql"
	"fmt"

	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/ntttrang/go-genai-slack-assistant/internal/service"
)

// TranslationRepositoryImpl implements service.TranslationRepository interface
type TranslationRepositoryImpl struct {
	db *sql.DB
}

// NewTranslationRepository creates a new translation repository instance
func NewTranslationRepository(db *sql.DB) service.TranslationRepository {
	return &TranslationRepositoryImpl{db: db}
}

func (tr *TranslationRepositoryImpl) Save(translation *model.Translation) error {
	query := `
		INSERT INTO translations (id, source_message_id, source_text, source_language, target_language, translated_text, hash, user_id, channel_id, created_at, ttl)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := tr.db.Exec(query,
		translation.ID,
		translation.SourceMessageID,
		translation.SourceText,
		translation.SourceLanguage,
		translation.TargetLanguage,
		translation.TranslatedText,
		translation.Hash,
		translation.UserID,
		translation.ChannelID,
		translation.CreatedAt,
		translation.TTL,
	)

	if err != nil {
		return fmt.Errorf("failed to save translation: %w", err)
	}

	return nil
}

func (tr *TranslationRepositoryImpl) GetByHash(hash string) (*model.Translation, error) {
	query := `
		SELECT id, source_message_id, source_text, source_language, target_language, translated_text, hash, user_id, channel_id, created_at, ttl
		FROM translations
		WHERE hash = ?
		LIMIT 1
	`

	translation := &model.Translation{}
	err := tr.db.QueryRow(query, hash).Scan(
		&translation.ID,
		&translation.SourceMessageID,
		&translation.SourceText,
		&translation.SourceLanguage,
		&translation.TargetLanguage,
		&translation.TranslatedText,
		&translation.Hash,
		&translation.UserID,
		&translation.ChannelID,
		&translation.CreatedAt,
		&translation.TTL,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get translation by hash: %w", err)
	}

	return translation, nil
}

func (tr *TranslationRepositoryImpl) GetByID(id string) (*model.Translation, error) {
	query := `
		SELECT id, source_message_id, source_text, source_language, target_language, translated_text, hash, user_id, channel_id, created_at, ttl
		FROM translations
		WHERE id = ?
	`

	translation := &model.Translation{}
	err := tr.db.QueryRow(query, id).Scan(
		&translation.ID,
		&translation.SourceMessageID,
		&translation.SourceText,
		&translation.SourceLanguage,
		&translation.TargetLanguage,
		&translation.TranslatedText,
		&translation.Hash,
		&translation.UserID,
		&translation.ChannelID,
		&translation.CreatedAt,
		&translation.TTL,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get translation by id: %w", err)
	}

	return translation, nil
}

func (tr *TranslationRepositoryImpl) GetByChannelID(channelID string, limit int) ([]*model.Translation, error) {
	query := `
		SELECT id, source_message_id, source_text, source_language, target_language, translated_text, hash, user_id, channel_id, created_at, ttl
		FROM translations
		WHERE channel_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := tr.db.Query(query, channelID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query translations: %w", err)
	}
	defer rows.Close()

	var translations []*model.Translation
	for rows.Next() {
		translation := &model.Translation{}
		err := rows.Scan(
			&translation.ID,
			&translation.SourceMessageID,
			&translation.SourceText,
			&translation.SourceLanguage,
			&translation.TargetLanguage,
			&translation.TranslatedText,
			&translation.Hash,
			&translation.UserID,
			&translation.ChannelID,
			&translation.CreatedAt,
			&translation.TTL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan translation: %w", err)
		}
		translations = append(translations, translation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return translations, nil
}
