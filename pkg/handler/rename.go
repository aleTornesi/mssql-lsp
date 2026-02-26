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

func (s *Server) handleTextDocumentRename(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.RenameParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	res, err := rename(f.Text, params)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func rename(text string, params lsp.RenameParams) (*lsp.WorkspaceEdit, error) {
	batchText, adjustedLine := parser.BatchAtLine(text, params.Position.Line)
	parsed, err := parser.Parse(batchText)
	if err != nil {
		return nil, err
	}

	pos := token.Pos{
		Line: adjustedLine,
		Col:  params.Position.Character,
	}

	batchStartLine := params.Position.Line - adjustedLine

	// Get the identifier on focus
	nodeWalker := parseutil.NewNodeWalker(parsed, pos)
	m := astutil.NodeMatcher{
		NodeTypes: []ast.NodeType{ast.TypeIdentifier},
	}
	currentVariable := nodeWalker.CurNodeBottomMatched(m)
	if currentVariable == nil {
		return nil, nil
	}

	name := currentVariable.String()

	// Try symbol-based rename for variables, CTEs, temp tables
	st := parseutil.ExtractSymbols(parsed)
	if sym := st.Lookup(name); sym != nil {
		refs := st.FindReferences(parsed, name)
		if len(refs) == 0 {
			return nil, nil
		}
		edits := make([]lsp.TextEdit, len(refs))
		nameLen := len(name)
		for i, ref := range refs {
			edits[i] = lsp.TextEdit{
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      ref.Line + batchStartLine,
						Character: ref.Col,
					},
					End: lsp.Position{
						Line:      ref.Line + batchStartLine,
						Character: ref.Col + nameLen,
					},
				},
				NewText: params.NewName,
			}
		}
		return &lsp.WorkspaceEdit{
			DocumentChanges: []lsp.TextDocumentEdit{
				{
					TextDocument: lsp.OptionalVersionedTextDocumentIdentifier{
						Version: 0,
						TextDocumentIdentifier: lsp.TextDocumentIdentifier{
							URI: params.TextDocument.URI,
						},
					},
					Edits: edits,
				},
			},
		}, nil
	}

	// Fall back to identifier-based rename
	idents, err := parseutil.ExtractIdenfiers(parsed, pos)
	if err != nil {
		return nil, err
	}

	renameTarget := []ast.Node{}
	for _, ident := range idents {
		if ident.String() == name {
			renameTarget = append(renameTarget, ident)
		}
	}
	if len(renameTarget) == 0 {
		return nil, nil
	}

	edits := make([]lsp.TextEdit, len(renameTarget))
	for i, target := range renameTarget {
		edit := lsp.TextEdit{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      target.Pos().Line + batchStartLine,
					Character: target.Pos().Col,
				},
				End: lsp.Position{
					Line:      target.End().Line + batchStartLine,
					Character: target.End().Col,
				},
			},
			NewText: params.NewName,
		}
		edits[i] = edit
	}

	res := &lsp.WorkspaceEdit{
		DocumentChanges: []lsp.TextDocumentEdit{
			{
				TextDocument: lsp.OptionalVersionedTextDocumentIdentifier{
					Version: 0,
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: params.TextDocument.URI,
					},
				},
				Edits: edits,
			},
		},
	}

	return res, nil
}
