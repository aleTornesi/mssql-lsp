package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/atornesi/tsql-ls/ast"
	"github.com/atornesi/tsql-ls/pkg/lsp"
	"github.com/atornesi/tsql-ls/parser"
	"github.com/atornesi/tsql-ls/parser/parseutil"
)

func (s *Server) handleTextDocumentDocumentSymbol(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DocumentSymbolParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	batches := parser.SplitBatches(f.Text)
	var symbols []lsp.DocumentSymbol

	for _, batch := range batches {
		parsed, err := parser.Parse(batch.Text)
		if err != nil {
			continue
		}

		// Extract declared symbols (variables, CTEs, temp tables)
		st := parseutil.ExtractSymbols(parsed)
		for _, sym := range st.Symbols {
			var kind lsp.SymbolKind
			var detail string
			switch sym.Kind {
			case parseutil.SymbolVariable:
				kind = lsp.SymbolKindVariable
				detail = sym.DataType
			case parseutil.SymbolCTE:
				kind = lsp.SymbolKindClass
				detail = "CTE"
			case parseutil.SymbolTempTable:
				kind = lsp.SymbolKindClass
				detail = "Temp Table"
			}
			r := lsp.Range{
				Start: lsp.Position{Line: sym.Pos.Line + batch.StartLine, Character: sym.Pos.Col},
				End:   lsp.Position{Line: sym.EndPos.Line + batch.StartLine, Character: sym.EndPos.Col},
			}
			symbols = append(symbols, lsp.DocumentSymbol{
				Name:           sym.Name,
				Detail:         detail,
				Kind:           kind,
				Range:          r,
				SelectionRange: r,
			})
		}

		// Extract block structures from AST
		collectBlockSymbols(parsed, batch.StartLine, &symbols)
	}

	if len(symbols) == 0 {
		return nil, nil
	}
	return symbols, nil
}

func collectBlockSymbols(tl ast.TokenList, lineOffset int, symbols *[]lsp.DocumentSymbol) {
	for _, node := range tl.GetTokens() {
		switch n := node.(type) {
		case *ast.BeginEnd:
			addBlockSymbol(n, "BEGIN...END", lsp.SymbolKindNamespace, lineOffset, symbols)
			collectBlockSymbols(n, lineOffset, symbols)
		case *ast.TryCatch:
			addBlockSymbol(n, "TRY...CATCH", lsp.SymbolKindNamespace, lineOffset, symbols)
			collectBlockSymbols(n, lineOffset, symbols)
		case *ast.IfStatement:
			addBlockSymbol(n, "IF", lsp.SymbolKindNamespace, lineOffset, symbols)
			collectBlockSymbols(n, lineOffset, symbols)
		case ast.TokenList:
			collectBlockSymbols(n, lineOffset, symbols)
		}
	}
}

func addBlockSymbol(node ast.Node, name string, kind lsp.SymbolKind, lineOffset int, symbols *[]lsp.DocumentSymbol) {
	start := node.Pos()
	end := node.End()
	if start.Line == end.Line && start.Col == end.Col {
		return
	}
	r := lsp.Range{
		Start: lsp.Position{Line: start.Line + lineOffset, Character: start.Col},
		End:   lsp.Position{Line: end.Line + lineOffset, Character: end.Col},
	}
	selRange := lsp.Range{
		Start: lsp.Position{Line: start.Line + lineOffset, Character: start.Col},
		End:   lsp.Position{Line: start.Line + lineOffset, Character: start.Col + len(name)},
	}
	*symbols = append(*symbols, lsp.DocumentSymbol{
		Name:           name,
		Kind:           kind,
		Range:          r,
		SelectionRange: selRange,
	})
}
