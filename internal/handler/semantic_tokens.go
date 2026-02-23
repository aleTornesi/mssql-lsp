package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/atornesi/tsql-ls/dialect"
	"github.com/atornesi/tsql-ls/internal/lsp"
	"github.com/atornesi/tsql-ls/parser"
	"github.com/atornesi/tsql-ls/token"
)

// Semantic token types - indices into the legend
const (
	semKeyword  = 0
	semType     = 1
	semFunction = 2
	semVariable = 3
	semString   = 4
	semNumber   = 5
	semComment  = 6
	semOperator = 7
)

var semanticTokenTypes = []string{
	"keyword",
	"type",
	"function",
	"variable",
	"string",
	"number",
	"comment",
	"operator",
}

var semanticTokenModifiers = []string{}

// Data type keywords for the "type" token type
var dataTypeKeywords = map[string]bool{
	"BIGINT": true, "BINARY": true, "BIT": true, "CHAR": true,
	"CHARACTER": true, "DATE": true, "DATETIME": true, "DATETIME2": true,
	"DATETIMEOFFSET": true, "DEC": true, "DECIMAL": true, "FLOAT": true,
	"IMAGE": true, "INT": true, "INTEGER": true, "MONEY": true,
	"NCHAR": true, "NTEXT": true, "NUMERIC": true, "NVARCHAR": true,
	"REAL": true, "SMALLDATETIME": true, "SMALLINT": true, "SMALLMONEY": true,
	"SQL_VARIANT": true, "TEXT": true, "TIME": true, "TIMESTAMP": true,
	"TINYINT": true, "UNIQUEIDENTIFIER": true, "VARBINARY": true,
	"VARCHAR": true, "XML": true, "CURSOR": true,
}

// Function names from the mssqlFunctions list
var functionKeywords map[string]bool

func init() {
	functionKeywords = make(map[string]bool)
	for _, fn := range dialect.DataBaseFunctions(dialect.DatabaseDriverMssql) {
		functionKeywords[strings.ToUpper(fn)] = true
	}
}

var semanticTokenCounter uint64

func (s *Server) handleTextDocumentSemanticTokensFull(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.SemanticTokensParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	data := encodeSemanticTokens(f.Text, nil)
	semanticTokenCounter++
	resultID := fmt.Sprintf("%d", semanticTokenCounter)
	s.semanticTokenCache[params.TextDocument.URI] = &cachedSemanticTokens{
		resultID: resultID,
		data:     data,
	}
	return lsp.SemanticTokens{ResultID: resultID, Data: data}, nil
}

func (s *Server) handleTextDocumentSemanticTokensDelta(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.SemanticTokensDeltaParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	newData := encodeSemanticTokens(f.Text, nil)
	semanticTokenCounter++
	resultID := fmt.Sprintf("%d", semanticTokenCounter)

	cached := s.semanticTokenCache[params.TextDocument.URI]
	s.semanticTokenCache[params.TextDocument.URI] = &cachedSemanticTokens{
		resultID: resultID,
		data:     newData,
	}

	// If no previous result or ID mismatch, return full tokens
	if cached == nil || cached.resultID != params.PreviousResultID {
		return lsp.SemanticTokens{ResultID: resultID, Data: newData}, nil
	}

	edits := computeSemanticTokenEdits(cached.data, newData)
	return lsp.SemanticTokensDelta{ResultID: resultID, Edits: edits}, nil
}

func computeSemanticTokenEdits(oldData, newData []uint32) []lsp.SemanticTokensEdit {
	// Find first difference
	minLen := len(oldData)
	if len(newData) < minLen {
		minLen = len(newData)
	}
	start := 0
	for start < minLen && oldData[start] == newData[start] {
		start++
	}

	// Find last difference from end
	oldEnd := len(oldData)
	newEnd := len(newData)
	for oldEnd > start && newEnd > start && oldData[oldEnd-1] == newData[newEnd-1] {
		oldEnd--
		newEnd--
	}

	if start == oldEnd && start == newEnd {
		return []lsp.SemanticTokensEdit{}
	}

	edit := lsp.SemanticTokensEdit{
		Start:       uint32(start),
		DeleteCount: uint32(oldEnd - start),
	}
	if newEnd > start {
		edit.Data = newData[start:newEnd]
	}
	return []lsp.SemanticTokensEdit{edit}
}

func (s *Server) handleTextDocumentSemanticTokensRange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.SemanticTokensRangeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	data := encodeSemanticTokens(f.Text, &params.Range)
	return lsp.SemanticTokens{Data: data}, nil
}

func encodeSemanticTokens(text string, rng *lsp.Range) []uint32 {
	batches := parser.SplitBatches(text)
	var data []uint32
	prevLine := 0
	prevStart := 0

	for _, batch := range batches {
		tokenizer := token.NewTokenizer(bytes.NewBufferString(batch.Text), &dialect.MSSQLDialect{})
		tokens, err := tokenizer.Tokenize()
		if err != nil {
			continue
		}

		for _, tok := range tokens {
			tokenType, ok := classifyToken(tok)
			if !ok {
				continue
			}

			line := tok.From.Line + batch.StartLine
			col := tok.From.Col
			length := tokenLength(tok)

			if length == 0 {
				continue
			}

			// Filter by range if specified
			if rng != nil {
				if line < rng.Start.Line || line > rng.End.Line {
					continue
				}
				if line == rng.Start.Line && col < rng.Start.Character {
					continue
				}
				if line == rng.End.Line && col >= rng.End.Character {
					continue
				}
			}

			deltaLine := line - prevLine
			deltaStart := col
			if deltaLine == 0 {
				deltaStart = col - prevStart
			}

			data = append(data, uint32(deltaLine), uint32(deltaStart), uint32(length), uint32(tokenType), 0)
			prevLine = line
			prevStart = col
		}
	}

	return data
}

func classifyToken(tok *token.Token) (int, bool) {
	switch tok.Kind {
	case token.SQLKeyword:
		w, ok := tok.Value.(*token.SQLWord)
		if !ok {
			return 0, false
		}
		upper := strings.ToUpper(w.Keyword)

		// @@ system variables
		if strings.HasPrefix(w.Value, "@@") {
			return semVariable, true
		}
		// @ variables
		if strings.HasPrefix(w.Value, "@") {
			return semVariable, true
		}

		// Quoted identifiers are not highlighted
		if w.QuoteStyle != 0 {
			return 0, false
		}

		// Data types
		if w.Kind != dialect.Unmatched && dataTypeKeywords[upper] {
			return semType, true
		}

		// Functions
		if functionKeywords[upper] {
			return semFunction, true
		}

		// Unmatched identifiers are not highlighted
		if w.Kind == dialect.Unmatched {
			return 0, false
		}

		// Keywords (DML/DDL/DCL/Matched)
		return semKeyword, true

	case token.SingleQuotedString, token.NationalStringLiteral:
		return semString, true
	case token.Number:
		return semNumber, true
	case token.Comment, token.MultilineComment:
		return semComment, true
	case token.Eq, token.Neq, token.Lt, token.Gt, token.LtEq, token.GtEq,
		token.Plus, token.Minus, token.Mult, token.Div, token.Mod:
		return semOperator, true
	default:
		return 0, false
	}
}

func tokenLength(tok *token.Token) int {
	switch v := tok.Value.(type) {
	case *token.SQLWord:
		return len(v.String())
	case string:
		switch tok.Kind {
		case token.Comment:
			return len(v) + 2 // --
		case token.MultilineComment:
			// Calculate actual rendered length including newlines
			rendered := "/*" + v + "*/"
			// For multiline, just return length of first line
			firstNewline := strings.IndexByte(rendered, '\n')
			if firstNewline >= 0 {
				return firstNewline
			}
			return len(rendered)
		default:
			return len(v)
		}
	}
	return 0
}
