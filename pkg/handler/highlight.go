package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/atornesi/tsql-ls/ast"
	"github.com/atornesi/tsql-ls/ast/astutil"
	"github.com/atornesi/tsql-ls/pkg/lsp"
	"github.com/atornesi/tsql-ls/parser"
	"github.com/atornesi/tsql-ls/parser/parseutil"
	"github.com/atornesi/tsql-ls/token"
)

func (s *Server) handleTextDocumentDocumentHighlight(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DocumentHighlightParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	return documentHighlight(f.Text, params)
}

func documentHighlight(text string, params lsp.DocumentHighlightParams) ([]lsp.DocumentHighlight, error) {
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

	// Symbol-based highlights
	st := parseutil.ExtractSymbols(parsed)
	if sym := st.Lookup(name); sym != nil {
		refs := st.FindReferences(parsed, name)
		if len(refs) == 0 {
			return nil, nil
		}
		var highlights []lsp.DocumentHighlight
		for _, ref := range refs {
			kind := lsp.DocumentHighlightKindRead
			if ref.Line == sym.Pos.Line && ref.Col == sym.Pos.Col {
				kind = lsp.DocumentHighlightKindWrite
			}
			highlights = append(highlights, lsp.DocumentHighlight{
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      ref.Line + batchStartLine,
						Character: ref.Col,
					},
					End: lsp.Position{
						Line:      ref.Line + batchStartLine,
						Character: ref.Col + len(name),
					},
				},
				Kind: kind,
			})
		}
		return highlights, nil
	}

	// Alias fallback
	aliases := parseutil.ExtractAliased(parsed)
	var defNode ast.Node
	for _, v := range aliases {
		alias, _ := v.(*ast.Aliased)
		if alias.AliasedName.String() == name {
			defNode = alias.AliasedName
			break
		}
	}

	idents, err := parseutil.ExtractIdenfiers(parsed, pos)
	if err != nil {
		return nil, err
	}

	var highlights []lsp.DocumentHighlight
	for _, ident := range idents {
		if ident.String() != name {
			continue
		}
		kind := lsp.DocumentHighlightKindRead
		if defNode != nil && ident.Pos().Line == defNode.Pos().Line && ident.Pos().Col == defNode.Pos().Col {
			kind = lsp.DocumentHighlightKindWrite
		}
		highlights = append(highlights, lsp.DocumentHighlight{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      ident.Pos().Line + batchStartLine,
					Character: ident.Pos().Col,
				},
				End: lsp.Position{
					Line:      ident.End().Line + batchStartLine,
					Character: ident.End().Col,
				},
			},
			Kind: kind,
		})
	}

	return highlights, nil
}
