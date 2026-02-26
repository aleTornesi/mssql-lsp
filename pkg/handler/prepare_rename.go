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

func (s *Server) handleTextDocumentPrepareRename(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.PrepareRenameParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	res, err := prepareRename(f.Text, params)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func prepareRename(text string, params lsp.PrepareRenameParams) (*lsp.PrepareRenameResult, error) {
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

	nodeWalker := parseutil.NewNodeWalker(parsed, pos)
	m := astutil.NodeMatcher{
		NodeTypes: []ast.NodeType{ast.TypeIdentifier},
	}
	currentNode := nodeWalker.CurNodeBottomMatched(m)
	if currentNode == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: "cannot rename this element"}
	}

	return &lsp.PrepareRenameResult{
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      currentNode.Pos().Line + batchStartLine,
				Character: currentNode.Pos().Col,
			},
			End: lsp.Position{
				Line:      currentNode.End().Line + batchStartLine,
				Character: currentNode.End().Col,
			},
		},
		Placeholder: currentNode.String(),
	}, nil
}
