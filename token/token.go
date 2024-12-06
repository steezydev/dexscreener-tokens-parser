package token

import "time"

type Token struct {
	PairID    string    `bson:"pairID"`
	Address   string    `bson:"address"`
	Symbol    string    `bson:"symbol"`
	Name      string    `bson:"name"`
	Price     float64   `bson:"price"`
	MarketCap float64   `bson:"marketCap"`
	Volume24h float64   `bson:"volume24h"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

// GetKey returns a unique identifier for the token
func GetKey(t Token) string {
	if t.Address != "" {
		return t.Address
	}
	return t.Symbol + ":" + t.Name
}

// RemoveDuplicates removes duplicate tokens from the slice
func RemoveDuplicates(tokens []Token) []Token {
	seen := make(map[string]bool)
	result := make([]Token, 0)

	for _, t := range tokens {
		key := GetKey(t)
		if !seen[key] {
			seen[key] = true
			result = append(result, t)
		}
	}
	return result
}
