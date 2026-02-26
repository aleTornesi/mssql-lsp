package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/aleTornesi/mssql-lsp/ast"
	"github.com/aleTornesi/mssql-lsp/ast/astutil"
	"github.com/aleTornesi/mssql-lsp/pkg/database"
	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
	"github.com/aleTornesi/mssql-lsp/parser"
	"github.com/aleTornesi/mssql-lsp/parser/parseutil"
	"github.com/aleTornesi/mssql-lsp/token"
)

func (s *Server) handleTypeDefinition(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.TypeDefinitionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	return typeDefinition(params.TextDocument.URI, f.Text, params)
}

func typeDefinition(uri, text string, params lsp.TypeDefinitionParams) (lsp.Definition, error) {
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
	st := parseutil.ExtractSymbols(parsed)
	sym := st.Lookup(name)
	if sym == nil || sym.Kind != parseutil.SymbolVariable || sym.DataType == "" {
		// Fall back to definition behavior
		defParams := lsp.DefinitionParams{
			TextDocumentPositionParams: params.TextDocumentPositionParams,
			WorkDoneProgressParams:     params.WorkDoneProgressParams,
			PartialResultParams:        params.PartialResultParams,
		}
		return definition(uri, text, defParams, nil)
	}

	// Find the type token position: walk tokens from symbol declaration
	toks := parseutil.FlattenTokens(parsed)
	for i, tok := range toks {
		if tok.From.Line == sym.Pos.Line && tok.From.Col == sym.Pos.Col {
			// Found the variable declaration token, skip WS to get to type
			j := parseutil.SkipWS(toks, i+1)
			if j < len(toks) {
				typeTok := toks[j]
				return []lsp.Location{
					{
						URI: uri,
						Range: lsp.Range{
							Start: lsp.Position{
								Line:      typeTok.From.Line + batchStartLine,
								Character: typeTok.From.Col,
							},
							End: lsp.Position{
								Line:      typeTok.To.Line + batchStartLine,
								Character: typeTok.To.Col,
							},
						},
					},
				}, nil
			}
		}
	}

	return nil, nil
}

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
