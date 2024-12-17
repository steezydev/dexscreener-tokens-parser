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

type RaydiumFetcher struct {
	client    *http.Client
	semaphore chan struct{}
}

func NewRaydium() *RaydiumFetcher {
	return &RaydiumFetcher{
		client:    &http.Client{Timeout: 10 * time.Second},
		semaphore: make(chan struct{}, 5),
	}
}

type raydiumResponse struct {
	Data struct {
		Data []struct {
			Id string `json:"id"`
		} `json:"data"`
	} `json:"data"`
}

func (f *RaydiumFetcher) FetchLPAddresses(tokens []token.Token) []token.Token {
	var wg sync.WaitGroup

	for i := range tokens {
		if tokens[i].Address == "" {
			log.Printf("Warning: Token %s:%s has no address to fetch LP", tokens[i].Symbol, tokens[i].Name)
			continue
		}

		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f.semaphore <- struct{}{}
			defer func() { <-f.semaphore }()

			lpAddress, err := f.fetchLPAddress(tokens[i].Address)
			if err != nil {
				log.Printf("Error fetching LP address for %s:%s: %v",
					tokens[i].Symbol, tokens[i].Name, err)
				return
			}

			tokens[i].LPAddress = lpAddress
			log.Printf("Found LP address %s for token %s:%s",
				lpAddress, tokens[i].Symbol, tokens[i].Name)

			time.Sleep(100 * time.Millisecond)
		}(i)
	}

	wg.Wait()
	return tokens
}

func (f *RaydiumFetcher) fetchLPAddress(tokenAddr string) (string, error) {
	url := fmt.Sprintf("https://api-v3.raydium.io/pools/info/mint?mint1=%s&mint2=So11111111111111111111111111111111111111112&poolType=all&poolSortField=liquidity&sortType=desc&pageSize=1000&page=1", tokenAddr)

	resp, err := f.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	var result raydiumResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode failed: %w", err)
	}

	if len(result.Data.Data) == 0 {
		return "", fmt.Errorf("no pool data found")
	}

	return result.Data.Data[0].Id, nil
}
