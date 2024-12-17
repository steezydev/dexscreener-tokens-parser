package scraper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gagliardetto/solana-go"
	"github.com/steezydev/dexscreener-tokens-parser/token"
)

type Scraper struct {
	apiKey string
}

func New(apiKey string) *Scraper {
	return &Scraper{
		apiKey: apiKey,
	}
}

func (s *Scraper) ScrapeTokens(pageNum int) ([]token.Token, error) {
	baseURL := "https://dexscreener.com/solana/raydium"
	if pageNum > 1 {
		baseURL = fmt.Sprintf("%s/page-%d", baseURL, pageNum)
	}

	params := url.Values{}
	params.Add("min24HTxns", "50")
	params.Add("minLiq", "100000")
	params.Add("minMarketCap", "5000000")
	params.Add("order", "desc")
	params.Add("rankBy", "marketCap")

	dexscreenerURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	zenrowsURL := url.URL{
		Scheme: "https",
		Host:   "api.zenrows.com",
		Path:   "/v1/",
	}

	zenrowsParams := url.Values{}
	zenrowsParams.Add("apikey", s.apiKey)
	zenrowsParams.Add("url", dexscreenerURL)
	zenrowsParams.Add("js_render", "true")
	zenrowsURL.RawQuery = zenrowsParams.Encode()

	doc, err := s.fetchDocument(zenrowsURL.String())
	if err != nil {
		return nil, err
	}

	return s.parseTokens(doc), nil
}

func (s *Scraper) fetchDocument(url string) (*goquery.Document, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response status: %s, body: %s", resp.Status, string(body))
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func (s *Scraper) parseTokens(doc *goquery.Document) []token.Token {
	var tokens []token.Token

	doc.Find(".ds-dex-table-row").Each(func(i int, s *goquery.Selection) {
		quoteToken := strings.TrimSpace(s.Find(".ds-dex-table-row-quote-token-symbol").Text())
		if quoteToken != "SOL" {
			return
		}

		t := token.Token{
			UpdatedAt: time.Now(),
		}

		if href, exists := s.Attr("href"); exists {
			t.PairID = strings.TrimPrefix(href, "/solana/")
		}

		t.Address = extractTokenAddress(s.Find(".ds-dex-table-row-col-token"))
		t.Symbol = strings.TrimSpace(s.Find(".ds-dex-table-row-base-token-symbol").Text())
		t.Name = strings.TrimSpace(s.Find(".ds-dex-table-row-base-token-name-text").Text())

		priceStr := s.Find(".ds-dex-table-row-col-price").Text()
		t.Price = parseNumericString(priceStr)

		mcapStr := s.Find(".ds-dex-table-row-col-market-cap").Text()
		t.MarketCap = parseNumericString(mcapStr)

		volumeStr := s.Find(".ds-dex-table-row-col-volume").Text()
		t.Volume24h = parseNumericString(volumeStr)

		tokens = append(tokens, t)
	})

	return tokens
}

func isValidSolanaAddress(address string) bool {
	_, err := solana.PublicKeyFromBase58(address)
	return err == nil
}

func extractTokenAddress(s *goquery.Selection) string {
	// Find the img element within the token icon div
	if img := s.Find(".ds-dex-table-row-token-icon-img"); img.Length() > 0 {
		if src, exists := img.Attr("src"); exists {
			if u, err := url.Parse(src); err == nil {
				// Extract filename without extension
				filename := path.Base(u.Path)
				address := strings.Split(filename, ".")[0] // Split by dot to remove everything after first period
				if isValidSolanaAddress(address) {
					return address
				}
			}
		}
	}
	return ""
}

func parseNumericString(s string) float64 {
	s = strings.TrimPrefix(s, "$")
	s = strings.TrimSuffix(s, "M")
	s = strings.TrimSuffix(s, "K")
	s = strings.TrimSuffix(s, "B")

	var multiplier float64 = 1
	if strings.HasSuffix(s, "M") {
		multiplier = 1000000
	} else if strings.HasSuffix(s, "K") {
		multiplier = 1000
	} else if strings.HasSuffix(s, "B") {
		multiplier = 1000000000
	}

	var val float64
	fmt.Sscanf(s, "%f", &val)
	return val * multiplier
}
