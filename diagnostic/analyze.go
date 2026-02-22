package diagnostic

import (
	"bytes"

	"github.com/atornesi/tsql-ls/dialect"
	"github.com/atornesi/tsql-ls/parser"
	"github.com/atornesi/tsql-ls/token"
)

// Analyze runs all diagnostic checks on a full T-SQL document.
func Analyze(text string) []Diagnostic {
	batches := parser.SplitBatches(text)
	var all []Diagnostic
	for _, batch := range batches {
		c := AnalyzeBatch(batch.Text)
		c.OffsetLines(batch.StartLine)
		all = append(all, c.Diagnostics...)
	}
	return all
}

// AnalyzeBatch runs lexer and structural checks on a single batch.
func AnalyzeBatch(text string) *Collector {
	c := &Collector{}

	// Lexer diagnostics
	src := bytes.NewBuffer([]byte(text))
	tokenizer := token.NewTokenizer(src, &dialect.MSSQLDialect{})
	tokenizer.Tokenize()

	for _, ld := range tokenizer.GetDiagnostics() {
		sev := Error
		if ld.Code == "TSQL003" {
			sev = Warning
		}
		c.Add(Diagnostic{
			From:     ld.From,
			To:       ld.To,
			Severity: sev,
			Message:  ld.Message,
			Code:     ld.Code,
		})
	}

	// AST structural checks (re-tokenizes internally via parser.Parse)
	parsed, err := parser.Parse(text)
	if err != nil {
		return c
	}
	CheckStructure(parsed, c)
	CheckSemantics(parsed, c)

	return c
}
