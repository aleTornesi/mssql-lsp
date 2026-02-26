package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/aleTornesi/mssql-lsp/ast"
	"github.com/aleTornesi/mssql-lsp/ast/astutil"
	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
	"github.com/aleTornesi/mssql-lsp/parser"
	"github.com/aleTornesi/mssql-lsp/parser/parseutil"
	"github.com/aleTornesi/mssql-lsp/token"
)

func (s *Server) handleTextDocumentSelectionRange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.SelectionRangeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	var ranges []lsp.SelectionRange
	for _, pos := range params.Positions {
		sr := selectionRange(f.Text, pos)
		ranges = append(ranges, sr)
	}
	return ranges, nil
}

func selectionRange(text string, pos lsp.Position) lsp.SelectionRange {
	batchText, adjustedLine := parser.BatchAtLine(text, pos.Line)
	parsed, err := parser.Parse(batchText)
	if err != nil {
		return leafRange(pos)
	}

	batchStartLine := pos.Line - adjustedLine
	tPos := token.Pos{Line: adjustedLine, Col: pos.Character}

	nw := parseutil.NewNodeWalker(parsed, tPos)
	nodes := collectAncestors(nw, tPos)

	if len(nodes) == 0 {
		return leafRange(pos)
	}

	// Build linked list from outermost to innermost
	var result *lsp.SelectionRange
	for _, node := range nodes {
		sr := &lsp.SelectionRange{
			Range: nodeToRange(node, batchStartLine),
		}
		if result != nil {
			sr.Parent = result
		}
		result = sr
	}

	if result == nil {
		return leafRange(pos)
	}
	return *result
}

func collectAncestors(nw *parseutil.NodeWalker, pos token.Pos) []ast.Node {
	var nodes []ast.Node
	for _, reader := range nw.Paths {
		nodes = append(nodes, reader.CurNode)
	}

	// Also check if the innermost node is a token list; if so, find the leaf token
	if len(nodes) > 0 {
		inner := nodes[len(nodes)-1]
		if _, ok := inner.(ast.TokenList); ok {
			// Try to find the leaf token at pos
			if leaf := findLeafToken(inner, pos); leaf != nil {
				nodes = append(nodes, leaf)
			}
		}
	}

	return nodes
}

func findLeafToken(node ast.Node, pos token.Pos) ast.Node {
	tl, ok := node.(ast.TokenList)
	if !ok {
		return nil
	}

	reader := astutil.NewNodeReader(tl)
	for reader.NextNode(false) {
		if reader.CurNodeEncloseIs(pos) {
			if _, isList := reader.CurNode.(ast.TokenList); !isList {
				return reader.CurNode
			}
			return findLeafToken(reader.CurNode, pos)
		}
	}
	return nil
}

func nodeToRange(node ast.Node, lineOffset int) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{
			Line:      node.Pos().Line + lineOffset,
			Character: node.Pos().Col,
		},
		End: lsp.Position{
			Line:      node.End().Line + lineOffset,
			Character: node.End().Col,
		},
	}
}

func leafRange(pos lsp.Position) lsp.SelectionRange {
	return lsp.SelectionRange{
		Range: lsp.Range{
			Start: pos,
			End:   pos,
		},
	}
}
