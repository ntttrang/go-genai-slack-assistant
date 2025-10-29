# Format Preservation in Translation

## Overview
This document explains how the translation system preserves text formatting (emojis, code blocks, links, line breaks) so that the output maintains the same format as the input.

## Problem
When translating text through AI services:
- Emojis might get corrupted or mistranslated
- Code blocks might be reformatted
- Links might be broken
- Line breaks might be lost
- Special formatting might be altered

**Solution**: Extract formatting before translation, translate clean text only, then restore formatting.

## Architecture

### FormatPreserver Service
Located in `internal/service/format_preserver.go`

#### Supported Format Types
1. **Emojis** - `:smile:`, `:wave:`, etc.
2. **Code Blocks** - `` `code` `` and ` ``` code ``` `
3. **Links** - URLs and Slack-specific links (`<@USER>`, `<#CHANNEL>`)
4. **Line Breaks** - Preserves newlines in multi-line messages

#### Workflow
```
Input Text
    ↓
Extract Emojis → Replace with EMOJI0, EMOJI1, ...
    ↓
Extract Code Blocks → Replace with CODEBLOCK0, CODEBLOCK1, ...
    ↓
Extract Links → Replace with LINK0, LINK1, ...
    ↓
Extract Line Breaks → Replace \n with LINEBREAK\n
    ↓
Clean Text (ready for translation)
    ↓
[Send to AI Translator]
    ↓
Translated Text (with placeholders)
    ↓
Restore Line Breaks ← Replace LINEBREAK\n with \n
    ↓
Restore Links ← Replace LINK0, LINK1, ... with original URLs
    ↓
Restore Code Blocks ← Replace CODEBLOCK0, CODEBLOCK1, ... with original code
    ↓
Restore Emojis ← Replace EMOJI0, EMOJI1, ... with original emojis
    ↓
Output Text (with formatting preserved)
```

## Integration in Translation Flow

The `TranslationUseCase.Translate()` method now:

1. **Extract Formatting** (Step 1)
   ```go
   preserver := NewFormatPreserver()
   textWithoutFormat := preserver.Extract(req.Text)
   ```

2. **Validate Cleaned Text** (Step 2)
   ```go
   inputValidation, err := tu.securityMiddleware.ValidateInput(textWithoutFormat)
   ```

3. **Translate Clean Text** (Step 6)
   ```go
   translatedText, err := tu.translator.Translate(sanitizedText, ...)
   ```

4. **Restore Formatting** (Step 8)
   ```go
   restoredTranslatedText := preserver.Restore(translatedText)
   ```

5. **Return Formatted Translation**
   ```go
   return response.Translation{
       OriginalText:   req.Text,
       TranslatedText: restoredTranslatedText,
       SourceLanguage: req.SourceLanguage,
       TargetLanguage: req.TargetLanguage,
   }, nil
   ```

## Examples

### Example 1: Emoji Preservation
**Input:**
```
:wave: Hello world :smile:
```

**During Translation:**
```
EMOJI0 Hello world EMOJI1
```

**Output:**
```
:wave: Hello world :smile:
```

### Example 2: Code Block Preservation
**Input:**
```
Use `npm start` to run this

```go
func main() {
    fmt.Println("Hello")
}
```
```

**During Translation:**
```
Use CODEBLOCK0 to run this

CODEBLOCK1
```

**Output:**
```
Use `npm start` to run this

```go
func main() {
    fmt.Println("Hello")
}
```
```

### Example 3: Combined Formatting
**Input:**
```
:wave: Hello!
Check https://example.com
See `config.js`
```

**During Translation:**
```
EMOJI0 Hello!
Check LINK0
See CODEBLOCK0
```

**Output:**
```
:wave: Hello!
Check https://example.com
See `config.js`
```

## Database & Cache

- **Database Storage**: Stores translations **without formatting** for consistency
  - Allows translations to be reused across different formatted versions
  - Reduces database size

- **Cache Retrieval**: Applies formatting when retrieving from cache
  - Format is restored before returning to user

## Performance Impact

- **Minimal overhead**: O(n) extraction and restoration for string replacements
- **Efficient pattern matching**: Uses compiled regex patterns
- **No additional API calls**: All processing happens locally

## Testing

Comprehensive tests in `internal/service/format_preserver_test.go` cover:
- Individual emoji preservation
- Code block preservation (inline and multiline)
- Link preservation (URLs and Slack-specific)
- Line break preservation
- Combined format preservation
- Edge cases and reset functionality

Run tests:
```bash
go test ./internal/service/... -v -run TestFormatPreserver
```

## Future Enhancements

Potential improvements:
- Slack formatting: `*bold*`, `_italic_`, `~strikethrough~`
- Numbered/bullet lists preservation
- Block quotes preservation
- User mentions (@user) special handling
- Channel mentions (#channel) special handling
