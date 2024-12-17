package fetcher

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/steezydev/dexscreener-tokens-parser/token"
)

type DexscreenerFetcher struct {
	client    *http.Client
	semaphore chan struct{}
}

func NewDexscreener() *DexscreenerFetcher {
	return &DexscreenerFetcher{
		client:    &http.Client{Timeout: 10 * time.Second},
		semaphore: make(chan struct{}, 5),
	}
}

type dexScreenerResponse struct {
	Pairs []struct {
		BaseToken struct {
			Address string `json:"address"`
			Name    string `json:"name"`
			Symbol  string `json:"symbol"`
		} `json:"baseToken"`
	} `json:"pairs"`
}

func (f *DexscreenerFetcher) FetchMissingAddresses(tokens []token.Token) []token.Token {
	var wg sync.WaitGroup

	for i := range tokens {
		if tokens[i].Address != "" {
			continue
		}

		if tokens[i].PairID == "" {
			log.Printf("Warning: Token %s:%s has no pair ID to fetch address", tokens[i].Symbol, tokens[i].Name)
			continue
		}

		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f.semaphore <- struct{}{}
			defer func() { <-f.semaphore }()

			address, err := f.fetchTokenAddress(tokens[i].PairID)
			if err != nil {
				log.Printf("Error fetching address for %s:%s (pair %s): %v",
					tokens[i].Symbol, tokens[i].Name, tokens[i].PairID, err)
				return
			}

			tokens[i].Address = address
			log.Printf("Found address %s for token %s:%s",
				address, tokens[i].Symbol, tokens[i].Name)

			time.Sleep(100 * time.Millisecond)
		}(i)
	}

	wg.Wait()
	return tokens
}

func (f *DexscreenerFetcher) fetchTokenAddress(pairID string) (string, error) {
	url := fmt.Sprintf("https://api.dexscreener.com/latest/dex/pairs/solana/%s", pairID)

	resp, err := f.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad response status: %s", resp.Status)
	}

	var result dexScreenerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if len(result.Pairs) > 0 {
		return result.Pairs[0].BaseToken.Address, nil
	}

	return "", fmt.Errorf("no pair data found")
}
