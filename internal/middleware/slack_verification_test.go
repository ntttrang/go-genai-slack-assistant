package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestVerifySlackSignatureGin_ValidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)

	signingSecret := "test-signing-secret"
	body := `{"type":"url_verification","challenge":"test"}`
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	hash := hmac.New(sha256.New, []byte(signingSecret))
	hash.Write([]byte(baseString))
	signature := "v0=" + hex.EncodeToString(hash.Sum(nil))

	r := gin.New()
	r.Use(VerifySlackSignatureGin(signingSecret))
	r.POST("/slack/events", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewBufferString(body))
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestVerifySlackSignatureGin_InvalidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)

	signingSecret := "test-signing-secret"
	body := `{"type":"url_verification","challenge":"test"}`
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	r := gin.New()
	r.Use(VerifySlackSignatureGin(signingSecret))
	r.POST("/slack/events", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewBufferString(body))
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", "v0=invalid-signature")
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid signature")
}

func TestVerifySlackSignatureGin_MissingTimestamp(t *testing.T) {
	gin.SetMode(gin.TestMode)

	signingSecret := "test-signing-secret"
	body := `{"type":"url_verification","challenge":"test"}`

	r := gin.New()
	r.Use(VerifySlackSignatureGin(signingSecret))
	r.POST("/slack/events", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewBufferString(body))
	req.Header.Set("X-Slack-Signature", "v0=abc123")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Missing signature headers")
}

func TestVerifySlackSignatureGin_MissingSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)

	signingSecret := "test-signing-secret"
	body := `{"type":"url_verification","challenge":"test"}`
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	r := gin.New()
	r.Use(VerifySlackSignatureGin(signingSecret))
	r.POST("/slack/events", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewBufferString(body))
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Missing signature headers")
}

func TestVerifySlackSignatureGin_TimestampTooOld(t *testing.T) {
	gin.SetMode(gin.TestMode)

	signingSecret := "test-signing-secret"
	body := `{"type":"url_verification","challenge":"test"}`
	oldTimestamp := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)

	baseString := fmt.Sprintf("v0:%s:%s", oldTimestamp, body)
	hash := hmac.New(sha256.New, []byte(signingSecret))
	hash.Write([]byte(baseString))
	signature := "v0=" + hex.EncodeToString(hash.Sum(nil))

	r := gin.New()
	r.Use(VerifySlackSignatureGin(signingSecret))
	r.POST("/slack/events", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewBufferString(body))
	req.Header.Set("X-Slack-Request-Timestamp", oldTimestamp)
	req.Header.Set("X-Slack-Signature", signature)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Request timestamp too old")
}

func TestVerifySlackSignatureGin_DifferentSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serverSecret := "server-secret"
	clientSecret := "client-secret"
	body := `{"type":"url_verification","challenge":"test"}`
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	hash := hmac.New(sha256.New, []byte(clientSecret))
	hash.Write([]byte(baseString))
	signature := "v0=" + hex.EncodeToString(hash.Sum(nil))

	r := gin.New()
	r.Use(VerifySlackSignatureGin(serverSecret))
	r.POST("/slack/events", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewBufferString(body))
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", signature)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid signature")
}

func TestVerifySlackSignatureGin_BodyReadability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	signingSecret := "test-signing-secret"
	body := `{"type":"url_verification","challenge":"test-body-content"}`
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	hash := hmac.New(sha256.New, []byte(signingSecret))
	hash.Write([]byte(baseString))
	signature := "v0=" + hex.EncodeToString(hash.Sum(nil))

	r := gin.New()
	r.Use(VerifySlackSignatureGin(signingSecret))
	r.POST("/slack/events", func(c *gin.Context) {
		var payload map[string]interface{}
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"challenge": payload["challenge"]})
	})

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewBufferString(body))
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "test-body-content")
}

func TestVerifySlackSignatureGin_ValidSignatureMultipleTimes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	signingSecret := "test-signing-secret"
	body := `{"type":"event_callback","event":{"type":"message"}}`
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	hash := hmac.New(sha256.New, []byte(signingSecret))
	hash.Write([]byte(baseString))
	signature := "v0=" + hex.EncodeToString(hash.Sum(nil))

	r := gin.New()
	r.Use(VerifySlackSignatureGin(signingSecret))
	r.POST("/slack/events", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/slack/events", bytes.NewBufferString(body))
		req.Header.Set("X-Slack-Request-Timestamp", timestamp)
		req.Header.Set("X-Slack-Signature", signature)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	}
}
