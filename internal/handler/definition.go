package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/atornesi/tsql-ls/ast"
	"github.com/atornesi/tsql-ls/ast/astutil"
	"github.com/atornesi/tsql-ls/internal/database"
	"github.com/atornesi/tsql-ls/internal/lsp"
	"github.com/atornesi/tsql-ls/parser"
	"github.com/atornesi/tsql-ls/parser/parseutil"
	"github.com/atornesi/tsql-ls/token"
)

func (s *Server) handleDefinition(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DefinitionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	return definition(params.TextDocument.URI, f.Text, params, s.worker.Cache())
}

func definition(uri, text string, params lsp.DefinitionParams, dbCache *database.DBCache) (lsp.Definition, error) {
	batchText, adjustedLine := parser.BatchAtLine(text, params.Position.Line)
	batchStartLine := params.Position.Line - adjustedLine
	pos := token.Pos{
		Line: adjustedLine,
		Col:  params.Position.Character + 1,
	}
	parsed, err := parser.Parse(batchText)
	if err != nil {
		return nil, err
	}

	nodeWalker := parseutil.NewNodeWalker(parsed, pos)
	m := astutil.NodeMatcher{
		NodeTypes: []ast.NodeType{ast.TypeIdentifier},
	}
	currentVariable := nodeWalker.CurNodeBottomMatched(m)
	if currentVariable == nil {
		return nil, nil
	}

	name := currentVariable.String()

	// Check symbol table (variables, CTEs, temp tables)
	st := parseutil.ExtractSymbols(parsed)
	if sym := st.Lookup(name); sym != nil {
		return []lsp.Location{
			{
				URI: uri,
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      sym.Pos.Line + batchStartLine,
						Character: sym.Pos.Col,
					},
					End: lsp.Position{
						Line:      sym.EndPos.Line + batchStartLine,
						Character: sym.EndPos.Col,
					},
				},
			},
		}, nil
	}

	// Fall back to alias lookup
	aliases := parseutil.ExtractAliased(parsed)
	if len(aliases) == 0 {
		return nil, nil
	}

	var define ast.Node
	for _, v := range aliases {
		alias, _ := v.(*ast.Aliased)
		if alias.AliasedName.String() == name {
			define = alias.AliasedName
			break
		}
	}

	if define == nil {
		return nil, nil
	}

	res := []lsp.Location{
		{
			URI: uri,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      define.Pos().Line + batchStartLine,
					Character: define.Pos().Col,
				},
				End: lsp.Position{
					Line:      define.End().Line + batchStartLine,
					Character: define.End().Col,
				},
			},
		},
	}

	return res, nil
}
