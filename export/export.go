package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/steezydev/dexscreener-tokens-parser/token"
)

// SaveToCSV saves the tokens to a CSV file with the given filename
func SaveToCSV(tokens []token.Token, filename string) error {
	// Ensure exports directory exists
	exportDir := "exports"
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return fmt.Errorf("error creating exports directory: %v", err)
	}

	// Create full file path
	fullPath := filepath.Join(exportDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"PairID",
		"Address",
		"LPAddress",
		"Symbol",
		"Name",
		"Price",
		"MarketCap",
		"Volume24h",
		"UpdatedAt",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %v", err)
	}

	// Write data
	for _, token := range tokens {
		row := []string{
			token.PairID,
			token.Address,
			token.LPAddress,
			token.Symbol,
			token.Name,
			strconv.FormatFloat(token.Price, 'f', -1, 64),
			strconv.FormatFloat(token.MarketCap, 'f', -1, 64),
			strconv.FormatFloat(token.Volume24h, 'f', -1, 64),
			token.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row: %v", err)
		}
	}

	return nil
}
