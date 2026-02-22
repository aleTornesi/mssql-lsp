package handler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/atornesi/tsql-ls/ast"
	"github.com/atornesi/tsql-ls/internal/lsp"
	"github.com/atornesi/tsql-ls/parser"
)

func (s *Server) handleTextDocumentCodeAction(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.CodeActionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	var actions []lsp.CodeAction
	uri := params.TextDocument.URI

	for _, diag := range params.Context.Diagnostics {
		if diag.Code == nil {
			continue
		}
		switch *diag.Code {
		case "TSQL011": // unreferenced CTE
			action := codeActionRemoveCTE(uri, f.Text, diag)
			if action != nil {
				actions = append(actions, *action)
			}
		case "TSQL004": // unclosed BEGIN
			actions = append(actions, codeActionAddEnd(uri, diag))
		case "TSQL007": // unclosed CASE
			actions = append(actions, codeActionAddEnd(uri, diag))
		}
	}

	if len(actions) == 0 {
		return []lsp.CodeAction{}, nil
	}
	return actions, nil
}

func codeActionRemoveCTE(uri, text string, diag lsp.Diagnostic) *lsp.CodeAction {
	// Find the CTE name from the diagnostic range and remove it from the WITH clause.
	// Simple approach: replace the diagnostic range text with empty string.
	// A full implementation would parse and remove the entire CTE definition,
	// but for now we provide a basic action.
	cteName := extractRangeText(text,
		diag.Range.Start.Line, diag.Range.Start.Character,
		diag.Range.End.Line, diag.Range.End.Character)
	if cteName == "" {
		return nil
	}
	return &lsp.CodeAction{
		Title:       fmt.Sprintf("Remove unreferenced CTE '%s'", cteName),
		Kind:        lsp.CodeActionKindQuickFix,
		Diagnostics: []lsp.Diagnostic{diag},
	}
}

func codeActionAddEnd(uri string, diag lsp.Diagnostic) lsp.CodeAction {
	// Insert END after the diagnostic range
	endPos := diag.Range.End
	insertLine := endPos.Line + 1
	return lsp.CodeAction{
		Title:       "Add missing END",
		Kind:        lsp.CodeActionKindQuickFix,
		Diagnostics: []lsp.Diagnostic{diag},
		Edit: &lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				uri: {
					{
						Range: lsp.Range{
							Start: lsp.Position{Line: insertLine, Character: 0},
							End:   lsp.Position{Line: insertLine, Character: 0},
						},
						NewText: "END\n",
					},
				},
			},
		},
	}
}

func (s *Server) handleWorkspaceExecuteCommand(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.ExecuteCommandParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("unsupported command: %v", params.Command)
}

func extractRangeText(text string, startLine, startChar, endLine, endChar int) string {
	writer := bytes.NewBufferString("")
	scanner := bufio.NewScanner(strings.NewReader(text))

	i := 0
	for scanner.Scan() {
		t := scanner.Text()
		if i >= startLine && i <= endLine {
			st, en := 0, len(t)

			if i == startLine {
				st = startChar
			}
			if i == endLine {
				en = endChar
			}

			writer.Write([]byte(t[st:en]))
			if i != endLine {
				writer.Write([]byte("\n"))
			}
		}
		i++
	}
	return writer.String()
}

func getStatements(text string) ([]*ast.Statement, error) {
	parsed, err := parser.Parse(text)
	if err != nil {
		return nil, err
	}

	var stmts []*ast.Statement
	for _, node := range parsed.GetTokens() {
		stmt, ok := node.(*ast.Statement)
		if !ok {
			return nil, fmt.Errorf("invalid type want Statement parsed %T", stmt)
		}
		stmts = append(stmts, stmt)
	}
	return stmts, nil
}
