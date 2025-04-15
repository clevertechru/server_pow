package service

import "github.com/clevertechru/server_pow/pkg/quotes"

type QuoteService struct{}

func NewQuoteService() *QuoteService {
	return &QuoteService{}
}

func (s *QuoteService) GetRandomQuote() string {
	return quotes.GetRandomQuote()
}
