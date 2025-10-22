package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func VerifySlackSignature(signingSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timestamp := r.Header.Get("X-Slack-Request-Timestamp")
			signature := r.Header.Get("X-Slack-Signature")

			if timestamp == "" || signature == "" {
				http.Error(w, "Missing signature headers", http.StatusUnauthorized)
				return
			}

			// Verify timestamp is not too old (within 5 minutes)
			reqTime, err := time.Parse(time.RFC3339, fmt.Sprintf("%sZ", timestamp))
			if err != nil {
				ts := time.Now().Unix()
				fmt.Sscanf(timestamp, "%d", &ts)
				reqTime = time.Unix(ts, 0)
			}

			if time.Since(reqTime) > 5*time.Minute {
				http.Error(w, "Request timestamp too old", http.StatusUnauthorized)
				return
			}

			// Read body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			// Restore body for handler
			r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

			// Verify signature
			baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(bodyBytes))
			hash := hmac.New(sha256.New, []byte(signingSecret))
			hash.Write([]byte(baseString))
			expectedSig := "v0=" + hex.EncodeToString(hash.Sum(nil))

			if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
				http.Error(w, "Invalid signature", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// VerifySlackSignatureGin is a Gin middleware for verifying Slack request signatures
func VerifySlackSignatureGin(signingSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		timestamp := c.GetHeader("X-Slack-Request-Timestamp")
		signature := c.GetHeader("X-Slack-Signature")

		if timestamp == "" || signature == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing signature headers"})
			return
		}

		// Verify timestamp is not too old (within 5 minutes)
		reqTime, err := time.Parse(time.RFC3339, fmt.Sprintf("%sZ", timestamp))
		if err != nil {
			ts := time.Now().Unix()
			fmt.Sscanf(timestamp, "%d", &ts)
			reqTime = time.Unix(ts, 0)
		}

		if time.Since(reqTime) > 5*time.Minute {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Request timestamp too old"})
			return
		}

		// Read body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Restore body for handler
		c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		// Verify signature
		baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(bodyBytes))
		hash := hmac.New(sha256.New, []byte(signingSecret))
		hash.Write([]byte(baseString))
		expectedSig := "v0=" + hex.EncodeToString(hash.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}

		c.Next()
	}
}
