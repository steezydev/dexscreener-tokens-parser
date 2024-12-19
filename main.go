package main

import (
	"fmt"
	"log"
	"time"

	"github.com/steezydev/dexscreener-tokens-parser/config"
	"github.com/steezydev/dexscreener-tokens-parser/export"
	"github.com/steezydev/dexscreener-tokens-parser/fetcher"
	"github.com/steezydev/dexscreener-tokens-parser/scraper"
	"github.com/steezydev/dexscreener-tokens-parser/storage"
	"github.com/steezydev/dexscreener-tokens-parser/token"
)

func main() {
	config := config.Get()
	apiKey := config.ZenrowsApiKey

	s := scraper.New(apiKey)

	store, err := storage.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	var allTokens []token.Token

	// Scrape all pages
	for page := 1; page <= 5; page++ {
		log.Printf("Scraping page %d...", page)
		tokens, err := s.ScrapeTokens(page)
		if err != nil {
			log.Printf("Error scraping page %d: %v", page, err)
			break
		}
		if len(tokens) == 0 {
			log.Printf("No tokens found on page %d, stopping", page)
			break
		}
		allTokens = append(allTokens, tokens...)
		log.Printf("Found %d tokens on page %d", len(tokens), page)
		time.Sleep(2 * time.Second)
	}

	// Remove duplicates
	log.Printf("Total tokens before deduplication: %d", len(allTokens))
	uniqueTokens := token.RemoveDuplicates(allTokens)
	log.Printf("Unique tokens after deduplication: %d", len(uniqueTokens))

	// Count tokens without addresses
	tokensWithoutAddress := 0
	for _, t := range uniqueTokens {
		if t.Address == "" {
			tokensWithoutAddress++
		}
	}
	log.Printf("Tokens without address: %d", tokensWithoutAddress)

	dexFetcher := fetcher.NewDexscreener()
	rayFetcher := fetcher.NewRaydium()

	// Fetch missing addresses
	if tokensWithoutAddress > 0 {
		log.Printf("Fetching missing token addresses...")
		uniqueTokens = dexFetcher.FetchMissingAddresses(uniqueTokens)
	}

	log.Printf("Fetching LP addresses...")
	uniqueTokens = rayFetcher.FetchLPAddresses(uniqueTokens)

	// Print tokens each on a new line
	for _, token := range uniqueTokens {
		fmt.Println(token)
	}

	if err := store.SaveTokens(uniqueTokens); err != nil {
		log.Printf("Error saving tokens: %v", err)
	}

	filename := fmt.Sprintf("tokens_%s.csv", time.Now().Format("2006-01-02_15-04-05"))
	if err := export.SaveToCSV(uniqueTokens, filename); err != nil {
		log.Printf("Error saving tokens to CSV: %v", err)
	} else {
		log.Printf("Successfully saved tokens to %s", filename)
	}

	log.Printf("Successfully processed %d tokens", len(uniqueTokens))
}
