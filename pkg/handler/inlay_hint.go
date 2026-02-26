package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
	"github.com/aleTornesi/mssql-lsp/parser"
	"github.com/aleTornesi/mssql-lsp/parser/parseutil"
)

func (s *Server) handleInlayHint(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.InlayHintParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	return inlayHints(f.Text, params)
}

func inlayHints(text string, params lsp.InlayHintParams) ([]lsp.InlayHint, error) {
	batches := parser.SplitBatches(text)
	var hints []lsp.InlayHint

	for _, batch := range batches {
		parsed, err := parser.Parse(batch.Text)
		if err != nil {
			continue
		}

		st := parseutil.ExtractSymbols(parsed)
		for _, sym := range st.Symbols {
			if sym.Kind != parseutil.SymbolVariable || sym.DataType == "" {
				continue
			}

			refs := st.FindReferences(parsed, sym.Name)
			for _, ref := range refs {
				// Skip declaration position
				if ref.Line == sym.Pos.Line && ref.Col == sym.Pos.Col {
					continue
				}

				line := ref.Line + batch.StartLine
				col := ref.Col + len(sym.Name)

				// Filter by requested range
				if line < params.Range.Start.Line || line > params.Range.End.Line {
					continue
				}
				if line == params.Range.Start.Line && col < params.Range.Start.Character {
					continue
				}
				if line == params.Range.End.Line && col > params.Range.End.Character {
					continue
				}

				hints = append(hints, lsp.InlayHint{
					Position:    lsp.Position{Line: line, Character: col},
					Label:       ": " + sym.DataType,
					Kind:        lsp.InlayHintKindType,
					PaddingLeft: true,
				})
			}
		}
	}

	if hints == nil {
		return []lsp.InlayHint{}, nil
	}
	return hints, nil
}
