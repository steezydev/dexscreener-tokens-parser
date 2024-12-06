# dexscreener-tokens-parser

Parser for Dexscreener token list

## Motivation

Dexscreener offers comprehensive token filtering features on it's frontend, which is unavailable though API. This script scrapes a filtered list of tokens from Dexscreener using Zenrow API and saves parsed data to the db.

## Features

- Scrapes tokens filtered by marketcap, liquidity and transaction count
- Tracks only SOL pairs
- Validates and fetches missing token addresses
- Daily token tracking with MongoDB storage

## Setup

1. Clone repository
2. Create `.env` file:

```env
ZENROWS_API_KEY=your_api_key
MONGODB_URL=your_mongo_url
MONGODB_DB=db_name
MONGODB_COLLECTION=collection_name
```

3. Run `go mod tidy`
4. Run `go run main.go`

## Token Schema

```json
{
  "date": "2024-12-06T00:00:00Z",
  "tokens": [
    {
      "pairId": "string",
      "address": "string",
      "symbol": "string",
      "name": "string",
      "marketCap": "number"
    }
  ],
  "totalTokens": "number",
  "scrapedAt": "2024-12-06T10:30:00Z"
}
```
