package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/atornesi/tsql-ls/ast"
	"github.com/atornesi/tsql-ls/internal/lsp"
	"github.com/atornesi/tsql-ls/parser"
	"github.com/atornesi/tsql-ls/parser/parseutil"
	"github.com/atornesi/tsql-ls/token"
)

func (s *Server) handleLinkedEditingRange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.LinkedEditingRangeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	return linkedEditingRange(f.Text, params)
}

func linkedEditingRange(text string, params lsp.LinkedEditingRangeParams) (*lsp.LinkedEditingRanges, error) {
	batchText, adjustedLine := parser.BatchAtLine(text, params.Position.Line)
	batchStartLine := params.Position.Line - adjustedLine
	parsed, err := parser.Parse(batchText)
	if err != nil {
		return nil, err
	}

	pos := token.Pos{
		Line: adjustedLine,
		Col:  params.Position.Character,
	}

	ranges := findBlockPairs(parsed, pos, batchStartLine)
	if len(ranges) == 0 {
		return nil, nil
	}
	return &lsp.LinkedEditingRanges{Ranges: ranges}, nil
}

func findBlockPairs(tl ast.TokenList, pos token.Pos, batchStartLine int) []lsp.Range {
	for _, node := range tl.GetTokens() {
		switch n := node.(type) {
		case *ast.BeginEnd:
			if r := beginEndPairs(n, pos, batchStartLine); r != nil {
				return r
			}
			if r := findBlockPairs(n, pos, batchStartLine); r != nil {
				return r
			}
		case *ast.TryCatch:
			if r := tryCatchPairs(n, pos, batchStartLine); r != nil {
				return r
			}
			if r := findBlockPairs(n, pos, batchStartLine); r != nil {
				return r
			}
		case *ast.SwitchCase:
			if r := caseEndPairs(n, pos, batchStartLine); r != nil {
				return r
			}
			if r := findBlockPairs(n, pos, batchStartLine); r != nil {
				return r
			}
		case ast.TokenList:
			if r := findBlockPairs(n, pos, batchStartLine); r != nil {
				return r
			}
		}
	}
	return nil
}

// beginEndPairs returns linked ranges for BEGIN/END if cursor is on either keyword.
func beginEndPairs(be *ast.BeginEnd, pos token.Pos, batchStartLine int) []lsp.Range {
	toks := parseutil.FlattenTokens(be)
	if len(toks) == 0 {
		return nil
	}

	// Find BEGIN token (first keyword)
	var beginTok, endTok *ast.SQLToken
	for _, tok := range toks {
		if parseutil.MatchKeyword(tok, "BEGIN") {
			if beginTok == nil {
				beginTok = tok
			}
		}
	}
	// Find END token (last keyword named END)
	for i := len(toks) - 1; i >= 0; i-- {
		if parseutil.MatchKeyword(toks[i], "END") {
			endTok = toks[i]
			break
		}
	}

	if beginTok == nil || endTok == nil {
		return nil
	}

	if isOnToken(pos, beginTok) || isOnToken(pos, endTok) {
		return []lsp.Range{
			tokenRange(beginTok, batchStartLine),
			tokenRange(endTok, batchStartLine),
		}
	}
	return nil
}

// tryCatchPairs returns linked ranges for BEGIN TRY/END TRY or BEGIN CATCH/END CATCH.
func tryCatchPairs(tc *ast.TryCatch, pos token.Pos, batchStartLine int) []lsp.Range {
	toks := parseutil.FlattenTokens(tc)

	// Collect keyword sequences: BEGIN TRY, END TRY, BEGIN CATCH, END CATCH
	type kwSpan struct {
		startTok *ast.SQLToken
		endTok   *ast.SQLToken
		label    string
	}
	var spans []kwSpan

	for i := 0; i < len(toks); i++ {
		if parseutil.MatchKeyword(toks[i], "BEGIN") || parseutil.MatchKeyword(toks[i], "END") {
			first := toks[i]
			j := parseutil.SkipWS(toks, i+1)
			if j < len(toks) && (parseutil.MatchKeyword(toks[j], "TRY") || parseutil.MatchKeyword(toks[j], "CATCH")) {
				label := strings.ToUpper(first.NoQuoteString()) + " " + strings.ToUpper(toks[j].NoQuoteString())
				spans = append(spans, kwSpan{startTok: first, endTok: toks[j], label: label})
				i = j
			}
		}
	}

	// Find pairs: BEGIN TRY <-> END TRY, BEGIN CATCH <-> END CATCH
	for _, pair := range []struct{ open, close string }{
		{"BEGIN TRY", "END TRY"},
		{"BEGIN CATCH", "END CATCH"},
	} {
		var openSpan, closeSpan *kwSpan
		for i := range spans {
			if spans[i].label == pair.open && openSpan == nil {
				openSpan = &spans[i]
			}
			if spans[i].label == pair.close {
				closeSpan = &spans[i]
			}
		}
		if openSpan == nil || closeSpan == nil {
			continue
		}
		if isOnToken(pos, openSpan.startTok) || isOnToken(pos, openSpan.endTok) ||
			isOnToken(pos, closeSpan.startTok) || isOnToken(pos, closeSpan.endTok) {
			return []lsp.Range{
				multiTokenRange(openSpan.startTok, openSpan.endTok, batchStartLine),
				multiTokenRange(closeSpan.startTok, closeSpan.endTok, batchStartLine),
			}
		}
	}
	return nil
}

// caseEndPairs returns linked ranges for CASE/END.
func caseEndPairs(sc *ast.SwitchCase, pos token.Pos, batchStartLine int) []lsp.Range {
	toks := parseutil.FlattenTokens(sc)
	if len(toks) == 0 {
		return nil
	}

	var caseTok, endTok *ast.SQLToken
	for _, tok := range toks {
		if parseutil.MatchKeyword(tok, "CASE") && caseTok == nil {
			caseTok = tok
		}
	}
	for i := len(toks) - 1; i >= 0; i-- {
		if parseutil.MatchKeyword(toks[i], "END") {
			endTok = toks[i]
			break
		}
	}

	if caseTok == nil || endTok == nil {
		return nil
	}

	if isOnToken(pos, caseTok) || isOnToken(pos, endTok) {
		return []lsp.Range{
			tokenRange(caseTok, batchStartLine),
			tokenRange(endTok, batchStartLine),
		}
	}
	return nil
}

func isOnToken(pos token.Pos, tok *ast.SQLToken) bool {
	if pos.Line < tok.From.Line || pos.Line > tok.To.Line {
		return false
	}
	if pos.Line == tok.From.Line && pos.Col < tok.From.Col {
		return false
	}
	if pos.Line == tok.To.Line && pos.Col > tok.To.Col {
		return false
	}
	return true
}

func tokenRange(tok *ast.SQLToken, batchStartLine int) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: tok.From.Line + batchStartLine, Character: tok.From.Col},
		End:   lsp.Position{Line: tok.To.Line + batchStartLine, Character: tok.To.Col},
	}
}

func multiTokenRange(first, last *ast.SQLToken, batchStartLine int) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: first.From.Line + batchStartLine, Character: first.From.Col},
		End:   lsp.Position{Line: last.To.Line + batchStartLine, Character: last.To.Col},
	}
}
