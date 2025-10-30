package service

import (
	"fmt"
	"regexp"
	"strings"
)

type FormatPreserver struct {
	emojis     map[string]string
	codeBlocks map[string]string
	links      map[string]string
	lineBreaks []int
	lists      map[string]string // stores list markers with indentation
	usernames  map[string]string // stores user ID to username mapping for mention conversion
}

func NewFormatPreserver() *FormatPreserver {
	return &FormatPreserver{
		emojis:     make(map[string]string),
		codeBlocks: make(map[string]string),
		links:      make(map[string]string),
		lists:      make(map[string]string),
		usernames:  make(map[string]string),
	}
}

// Extract preserves formatting by replacing patterns with placeholders
func (fp *FormatPreserver) Extract(text string) string {
	// 1. Extract list markers with indentation (before other extractions)
	text = fp.extractLists(text)
	
	// 2. Extract code blocks (backticks)
	text = fp.extractCodeBlocks(text)
	
	// 3. Extract links
	text = fp.extractLinks(text)
	
	// 4. Extract emoji codes
	text = fp.extractEmojis(text)
	
	// 5. Preserve line breaks as placeholders
	text = fp.extractLineBreaks(text)
	
	return text
}

// Restore applies all formatting back to translated text
func (fp *FormatPreserver) Restore(text string) string {
	return fp.RestoreWithOptions(text, false)
}

// RestoreWithOptions applies all formatting back with option to convert user mentions to plain text
func (fp *FormatPreserver) RestoreWithOptions(text string, convertUserMentions bool) string {
	// 1. Restore line breaks
	text = fp.restoreLineBreaks(text)
	
	// 2. Restore emoji codes
	text = fp.restoreEmojis(text)
	
	// 3. Restore links (optionally converting user mentions to plain text)
	text = fp.restoreLinksWithOptions(text, convertUserMentions)
	
	// 4. Restore code blocks
	text = fp.restoreCodeBlocks(text)
	
	// 5. Restore list markers
	text = fp.restoreLists(text)
	
	return text
}

func (fp *FormatPreserver) extractLists(text string) string {
	// Match bullet points (* or -) and numbered lists with optional indentation
	// Pattern: optional spaces, then (* or - or digit.), then space, then content
	listPattern := regexp.MustCompile(`^(\s*)([*\-]\s|\d+\.\s)(.*)$`)
	
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if match := listPattern.FindStringSubmatch(line); match != nil {
			// match[1] = indentation (spaces)
			// match[2] = list marker (* or - or digit.)
			// match[3] = content
			
			indentation := match[1]
			marker := match[2]
			content := match[3]
			
			// Create placeholder for the entire list line
			placeholder := fmt.Sprintf("LIST%d", len(fp.lists))
			// Store the indentation + marker for restoration
			fp.lists[placeholder] = indentation + marker
			
			// Replace line with placeholder + content
			lines[i] = placeholder + content
		}
	}
	
	return strings.Join(lines, "\n")
}

func (fp *FormatPreserver) extractCodeBlocks(text string) string {
	// Match single backticks `code` and triple backticks ```code```
	codePattern := regexp.MustCompile("```[\\s\\S]*?```|`[^`]*`")
	
	return codePattern.ReplaceAllStringFunc(text, func(match string) string {
		placeholder := fmt.Sprintf("CODEBLOCK%d", len(fp.codeBlocks))
		fp.codeBlocks[placeholder] = match
		return placeholder
	})
}

func (fp *FormatPreserver) extractLinks(text string) string {
	// Match URLs and Slack links <http://...> and <@USER> mentions
	linkPattern := regexp.MustCompile(`<[^>]+>|https?://[^\s]+`)
	
	return linkPattern.ReplaceAllStringFunc(text, func(match string) string {
		placeholder := fmt.Sprintf("LINK%d", len(fp.links))
		fp.links[placeholder] = match
		return placeholder
	})
}

func (fp *FormatPreserver) extractEmojis(text string) string {
	// Match emoji codes like :smile: :wave:
	emojiPattern := regexp.MustCompile(`:[a-zA-Z0-9_-]+:`)
	
	return emojiPattern.ReplaceAllStringFunc(text, func(match string) string {
		placeholder := fmt.Sprintf("EMOJI%d", len(fp.emojis))
		fp.emojis[placeholder] = match
		return placeholder
	})
}

func (fp *FormatPreserver) extractLineBreaks(text string) string {
	// Replace newlines with placeholder to preserve structure
	return strings.ReplaceAll(text, "\n", "LINEBREAK")
}

func (fp *FormatPreserver) restoreLineBreaks(text string) string {
	return strings.ReplaceAll(text, "LINEBREAK", "\n")
}

func (fp *FormatPreserver) restoreEmojis(text string) string {
	result := text
	for placeholder, emoji := range fp.emojis {
		result = strings.ReplaceAll(result, placeholder, emoji)
	}
	return result
}

func (fp *FormatPreserver) restoreLinks(text string) string {
	return fp.restoreLinksWithOptions(text, false)
}

func (fp *FormatPreserver) restoreLinksWithOptions(text string, convertUserMentions bool) string {
	result := text
	userMentionPattern := regexp.MustCompile(`<@(U[A-Z0-9]+)>`)
	
	for placeholder, link := range fp.links {
		if convertUserMentions && userMentionPattern.MatchString(link) {
			matches := userMentionPattern.FindStringSubmatch(link)
			if len(matches) > 1 {
				userID := matches[1]
				username := fp.usernames[userID]
				if username == "" {
					username = userID
				}
				result = strings.ReplaceAll(result, placeholder, username)
			} else {
				result = strings.ReplaceAll(result, placeholder, link)
			}
		} else {
			result = strings.ReplaceAll(result, placeholder, link)
		}
	}
	
	return strings.TrimSpace(result)
}

func (fp *FormatPreserver) restoreCodeBlocks(text string) string {
	result := text
	for placeholder, code := range fp.codeBlocks {
		result = strings.ReplaceAll(result, placeholder, code)
	}
	return result
}

func (fp *FormatPreserver) restoreLists(text string) string {
	result := text
	for placeholder, marker := range fp.lists {
		result = strings.ReplaceAll(result, placeholder, marker)
	}
	return result
}

// SetUsernameMappings stores the user ID to username mapping for mention conversion
func (fp *FormatPreserver) SetUsernameMappings(mappings map[string]string) {
	fp.usernames = mappings
}

// ExtractUserIDsFromText extracts all user IDs from Slack mentions in the text
func (fp *FormatPreserver) ExtractUserIDsFromText(text string) []string {
	userMentionPattern := regexp.MustCompile(`<@(U[A-Z0-9]+)>`)
	matches := userMentionPattern.FindAllStringSubmatch(text, -1)
	
	userIDs := make([]string, 0)
	seen := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) > 1 {
			userID := match[1]
			if !seen[userID] {
				userIDs = append(userIDs, userID)
				seen[userID] = true
			}
		}
	}
	
	return userIDs
}

// Reset clears all stored patterns for reuse
func (fp *FormatPreserver) Reset() {
	fp.emojis = make(map[string]string)
	fp.codeBlocks = make(map[string]string)
	fp.links = make(map[string]string)
	fp.lists = make(map[string]string)
	fp.usernames = make(map[string]string)
}
