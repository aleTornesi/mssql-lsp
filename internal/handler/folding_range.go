package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/atornesi/tsql-ls/ast"
	"github.com/atornesi/tsql-ls/dialect"
	"github.com/atornesi/tsql-ls/internal/lsp"
	"github.com/atornesi/tsql-ls/parser"
	"github.com/atornesi/tsql-ls/token"
)

func (s *Server) handleTextDocumentFoldingRange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.FoldingRangeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	batches := parser.SplitBatches(f.Text)
	var ranges []lsp.FoldingRange

	for _, batch := range batches {
		parsed, err := parser.Parse(batch.Text)
		if err != nil {
			continue
		}
		collectFoldingRanges(parsed, batch.StartLine, &ranges)
	}

	// Collect multiline comments from tokens
	collectCommentFoldingRanges(f.Text, &ranges)

	if len(ranges) == 0 {
		return nil, nil
	}
	return ranges, nil
}

func collectFoldingRanges(tl ast.TokenList, lineOffset int, ranges *[]lsp.FoldingRange) {
	for _, node := range tl.GetTokens() {
		switch n := node.(type) {
		case *ast.BeginEnd:
			addFoldingRange(n, lineOffset, lsp.FoldingRangeKindRegion, ranges)
			collectFoldingRanges(n, lineOffset, ranges)
		case *ast.TryCatch:
			addFoldingRange(n, lineOffset, lsp.FoldingRangeKindRegion, ranges)
			collectFoldingRanges(n, lineOffset, ranges)
		case *ast.IfStatement:
			addFoldingRange(n, lineOffset, lsp.FoldingRangeKindRegion, ranges)
			collectFoldingRanges(n, lineOffset, ranges)
		case *ast.SwitchCase:
			addFoldingRange(n, lineOffset, lsp.FoldingRangeKindRegion, ranges)
			collectFoldingRanges(n, lineOffset, ranges)
		case *ast.Parenthesis:
			start := n.Pos()
			end := n.End()
			if start.Line != end.Line {
				addFoldingRange(n, lineOffset, lsp.FoldingRangeKindRegion, ranges)
			}
			collectFoldingRanges(n, lineOffset, ranges)
		case ast.TokenList:
			collectFoldingRanges(n, lineOffset, ranges)
		}
	}
}

func addFoldingRange(node ast.Node, lineOffset int, kind lsp.FoldingRangeKind, ranges *[]lsp.FoldingRange) {
	start := node.Pos()
	end := node.End()
	if start.Line == end.Line {
		return
	}
	k := kind
	*ranges = append(*ranges, lsp.FoldingRange{
		StartLine: start.Line + lineOffset,
		EndLine:   end.Line + lineOffset,
		Kind:      &k,
	})
}

func collectCommentFoldingRanges(text string, ranges *[]lsp.FoldingRange) {
	tokenizer := token.NewTokenizer(bytes.NewBufferString(text), &dialect.MSSQLDialect{})
	tokens, err := tokenizer.Tokenize()
	if err != nil {
		return
	}
	for _, tok := range tokens {
		if tok.Kind == token.MultilineComment && tok.From.Line != tok.To.Line {
			k := lsp.FoldingRangeKindComment
			*ranges = append(*ranges, lsp.FoldingRange{
				StartLine: tok.From.Line,
				EndLine:   tok.To.Line,
				Kind:      &k,
			})
		}
	}
}
