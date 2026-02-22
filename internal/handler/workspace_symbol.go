package handler

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/atornesi/tsql-ls/internal/lsp"
	"github.com/atornesi/tsql-ls/parser"
	"github.com/atornesi/tsql-ls/parser/parseutil"
)

func (s *Server) handleWorkspaceSymbol(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.WorkspaceSymbolParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	query := strings.ToUpper(params.Query)
	var symbols []lsp.SymbolInformation

	for uri, f := range s.files {
		batches := parser.SplitBatches(f.Text)
		for _, batch := range batches {
			parsed, err := parser.Parse(batch.Text)
			if err != nil {
				continue
			}

			st := parseutil.ExtractSymbols(parsed)
			for _, sym := range st.Symbols {
				if query != "" && !strings.Contains(strings.ToUpper(sym.Name), query) {
					continue
				}
				var kind lsp.SymbolKind
				switch sym.Kind {
				case parseutil.SymbolVariable:
					kind = lsp.SymbolKindVariable
				case parseutil.SymbolCTE:
					kind = lsp.SymbolKindClass
				case parseutil.SymbolTempTable:
					kind = lsp.SymbolKindClass
				}
				r := lsp.Range{
					Start: lsp.Position{Line: sym.Pos.Line + batch.StartLine, Character: sym.Pos.Col},
					End:   lsp.Position{Line: sym.EndPos.Line + batch.StartLine, Character: sym.EndPos.Col},
				}
				symbols = append(symbols, lsp.SymbolInformation{
					Name: sym.Name,
					Kind: kind,
					Location: lsp.Location{
						URI:   uri,
						Range: r,
					},
				})
			}
		}
	}

	if len(symbols) == 0 {
		return nil, nil
	}
	return symbols, nil
}
