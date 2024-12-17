package storage

import (
	"context"
	"time"

	"github.com/steezydev/dexscreener-tokens-parser/config"
	"github.com/steezydev/dexscreener-tokens-parser/token"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TokenEntry struct {
	PairID    string  `bson:"pairId"`
	Address   string  `bson:"address"`
	LPAddress string  `bson:"lpAddress"`
	Symbol    string  `bson:"symbol"`
	Name      string  `bson:"name"`
	MarketCap float64 `bson:"marketCap"`
}

type DailyTokens struct {
	Date        time.Time    `bson:"date"`
	Tokens      []TokenEntry `bson:"tokens"`
	TotalTokens int          `bson:"totalTokens"`
	ScrapedAt   time.Time    `bson:"scrapedAt"`
}

type Storage struct {
	client *mongo.Client
	coll   *mongo.Collection
}

func New(cfg *config.Config) (*Storage, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURL))
	if err != nil {
		return nil, err
	}

	coll := client.Database(cfg.MongoDBName).Collection(cfg.MongoCollection)
	return &Storage{client: client, coll: coll}, nil
}

func (s *Storage) Close() error {
	return s.client.Disconnect(context.Background())
}

func (s *Storage) SaveTokens(tokens []token.Token) error {
	entries := make([]TokenEntry, 0, len(tokens))
	for _, t := range tokens {
		if t.Address == "" || t.PairID == "" {
			continue
		}
		entries = append(entries, TokenEntry{
			PairID:    t.PairID,
			Address:   t.Address,
			LPAddress: t.LPAddress,
			Symbol:    t.Symbol,
			Name:      t.Name,
			MarketCap: t.MarketCap,
		})
	}

	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	daily := DailyTokens{
		Date:        date,
		Tokens:      entries,
		TotalTokens: len(entries),
		ScrapedAt:   now,
	}

	_, err := s.coll.UpdateOne(
		context.Background(),
		bson.M{"date": date},
		bson.M{"$set": daily},
		options.Update().SetUpsert(true),
	)
	return err
}
