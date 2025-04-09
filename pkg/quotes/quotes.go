package quotes

import (
	"math/rand"
	"time"
)

var quotes = []string{
	"The only way to do great work is to love what you do. - Steve Jobs",
	"Stay hungry, stay foolish. - Steve Jobs",
	"Code is like humor. When you have to explain it, it's bad. - Cory House",
	"First, solve the problem. Then, write the code. - John Johnson",
	"Experience is the name everyone gives to their mistakes. - Oscar Wilde",
	"Programming isn't about what you know; it's about what you can figure out. - Chris Pine",
	"The only way to learn a new programming language is by writing programs in it. - Dennis Ritchie",
	"Talk is cheap. Show me the code. - Linus Torvalds",
	"Programming is the art of telling another human what one wants the computer to do. - Donald Knuth",
	"Clean code always looks like it was written by someone who cares. - Robert C. Martin",
}

func GetRandomQuote() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return quotes[r.Intn(len(quotes))]
}
