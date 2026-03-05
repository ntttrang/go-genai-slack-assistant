package goldprice

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	worldPriceURL    = "https://giavang.org/the-gioi/"
	domesticPriceURL = "https://giavang.org/trong-nuoc/"

	httpTimeout = 15 * time.Second
)

// DomesticPrice holds buy/sell prices for a single brand.
type DomesticPrice struct {
	Buy  string
	Sell string
}

// GoldPriceData holds all scraped gold price information.
type GoldPriceData struct {
	WorldPrice  string
	WorldChange string
	Domestic    map[string]DomesticPrice
}

// Scraper fetches gold price data from giavang.org.
type Scraper struct {
	httpClient *http.Client
}

// NewScraper creates a new Scraper with a sensible HTTP timeout.
func NewScraper() *Scraper {
	return &Scraper{
		httpClient: &http.Client{Timeout: httpTimeout},
	}
}

// Scrape fetches world and domestic gold prices concurrently.
func (s *Scraper) Scrape() (*GoldPriceData, error) {
	data := &GoldPriceData{
		Domestic: make(map[string]DomesticPrice),
	}

	var wg sync.WaitGroup
	errs := make([]error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		price, change, err := s.scrapeWorldPrice()
		if err != nil {
			errs[0] = fmt.Errorf("world price: %w", err)
			return
		}
		data.WorldPrice = price
		data.WorldChange = change
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		domestic, err := s.scrapeDomesticPrices()
		if err != nil {
			errs[1] = fmt.Errorf("domestic price: %w", err)
			return
		}
		data.Domestic = domestic
	}()

	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return data, err
		}
	}

	return data, nil
}

// scrapeWorldPrice parses the world gold price and percentage change from giavang.org/the-gioi/.
func (s *Scraper) scrapeWorldPrice() (price, change string, err error) {
	req, err := http.NewRequest(http.MethodGet, worldPriceURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoldPriceBot/1.0)")
	req.Header.Set("Accept-Language", "vi-VN,vi;q=0.9")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", err
	}

	// The page has an h1 containing the XAU/USD heading; the price and change
	// appear in the first visible text nodes / spans after it.
	// We look for the pattern: a large price number followed by a change line.
	// Structure observed: <div class="..."> 5,161.58 USD <span>38.11USD(0.74%)</span>
	// Try to locate the price container by scanning common wrappers.
	doc.Find("h1, h2, .price, .gold-price, [class*='price'], [class*='xau']").Each(func(_ int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		if price == "" && strings.Contains(text, "USD") {
			// Extract numeric price token
			for _, part := range strings.Fields(text) {
				part = strings.ReplaceAll(part, ",", "")
				if isNumeric(part) {
					price = strings.Fields(text)[0]
					break
				}
			}
		}
	})

	// Broader fallback: scan entire body text for price pattern.
	if price == "" {
		doc.Find("body").Each(func(_ int, sel *goquery.Selection) {
			text := sel.Text()
			price, change = extractWorldPriceFromText(text)
		})
	} else {
		// Extract change from same page body text.
		bodyText := doc.Find("body").Text()
		_, change = extractWorldPriceFromText(bodyText)
	}

	if price == "" {
		return "N/A", "N/A", nil
	}
	if change == "" {
		change = "N/A"
	}

	return price, change, nil
}

// scrapeDomesticPrices parses domestic gold prices for TP. Hồ Chí Minh from giavang.org/trong-nuoc/.
func (s *Scraper) scrapeDomesticPrices() (map[string]DomesticPrice, error) {
	result := make(map[string]DomesticPrice)

	req, err := http.NewRequest(http.MethodGet, domesticPriceURL, nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoldPriceBot/1.0)")
	req.Header.Set("Accept-Language", "vi-VN,vi;q=0.9")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return result, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return result, err
	}

	// Target brands to extract for TPHCM region.
	targetBrands := map[string]bool{
		"SJC":       true,
		"PNJ":       true,
		"DOJI":      true,
		"Mi Hồng":   true,
		"Ngọc Thẩm": true,
	}

	// The comparison table has rows: Khu vực | Hệ thống (brand) | Mua vào | Bán ra.
	// Rows for TPHCM have rowspan on the first cell; subsequent rows in the group
	// only have 3 cells (brand, buy, sell). We track whether we're inside the TPHCM group.
	inTPHCM := false

	doc.Find("table tr").Each(func(_ int, row *goquery.Selection) {
		cells := row.Find("td")
		count := cells.Length()

		if count == 0 {
			return
		}

		firstCellText := strings.TrimSpace(cells.Eq(0).Text())

		// A row with 4 cells typically starts a new region group.
		if count >= 4 {
			regionText := firstCellText
			inTPHCM = strings.Contains(regionText, "Hồ Chí Minh") ||
				strings.Contains(regionText, "TP.HCM") ||
				strings.Contains(regionText, "TPHCM")

			if inTPHCM {
				brand := strings.TrimSpace(cells.Eq(1).Text())
				buy := strings.TrimSpace(cells.Eq(2).Text())
				sell := strings.TrimSpace(cells.Eq(3).Text())
				if targetBrands[brand] {
					result[brand] = DomesticPrice{Buy: formatPrice(buy), Sell: formatPrice(sell)}
				}
			}
			return
		}

		// A row with 3 cells is a continuation of the current region group.
		if count == 3 && inTPHCM {
			brand := strings.TrimSpace(cells.Eq(0).Text())
			buy := strings.TrimSpace(cells.Eq(1).Text())
			sell := strings.TrimSpace(cells.Eq(2).Text())
			if targetBrands[brand] {
				result[brand] = DomesticPrice{Buy: formatPrice(buy), Sell: formatPrice(sell)}
			}
		}
	})

	return result, nil
}

// extractWorldPriceFromText attempts to find the first "number USD" pattern
// and the first percentage change in the given text blob.
func extractWorldPriceFromText(text string) (price, change string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Look for a line containing a price like "5,161.58 USD"
		if price == "" && strings.Contains(line, "USD") {
			fields := strings.Fields(line)
			for i, f := range fields {
				cleaned := strings.ReplaceAll(f, ",", "")
				if isNumericFloat(cleaned) && i+1 < len(fields) && fields[i+1] == "USD" {
					price = f
					break
				}
			}
		}
		// Look for percentage change: "tăng X%" or "giảm X%" or "(X%)"
		if change == "" {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "tăng") || strings.Contains(lower, "giảm") {
				// Extract the direction + percent
				direction := ""
				if strings.Contains(lower, "tăng") {
					direction = "Tăng"
				} else {
					direction = "Giảm"
				}
				for _, f := range strings.Fields(line) {
					if strings.Contains(f, "%") {
						change = direction + " " + f
						break
					}
				}
			}
		}
		if price != "" && change != "" {
			break
		}
	}
	return price, change
}

// formatPrice cleans up a price string (e.g. "181.700" → "181.700").
func formatPrice(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "N/A"
	}
	return raw
}

// isNumeric returns true if the string contains only digits and dots/commas.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c != '.' && c != ',' && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

// isNumericFloat returns true if s looks like a floating-point number.
func isNumericFloat(s string) bool {
	if s == "" {
		return false
	}
	dotSeen := false
	for i, c := range s {
		if c == '.' {
			if dotSeen {
				return false
			}
			dotSeen = true
		} else if c == '-' && i == 0 {
			continue
		} else if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
