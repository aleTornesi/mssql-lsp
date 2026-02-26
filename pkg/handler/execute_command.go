package handler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/aleTornesi/mssql-lsp/ast"
	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
	"github.com/aleTornesi/mssql-lsp/parser"
	"github.com/aleTornesi/mssql-lsp/parser/parseutil"
	"github.com/aleTornesi/mssql-lsp/token"
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
		case "TSQL005": // unclosed TRY/CATCH
			actions = append(actions, codeActionAddEndCatch(uri, diag))
		case "TSQL006": // unclosed paren
			actions = append(actions, codeActionAddCloseParen(uri, diag))
		case "TSQL007": // unclosed CASE
			actions = append(actions, codeActionAddEnd(uri, diag))
		case "TSQL010": // duplicate variable
			actions = append(actions, codeActionRemoveDuplicateDecl(uri, f.Text, diag))
		}
	}

	// Selection-based refactoring: surround with TRY/CATCH
	if params.Range.Start != params.Range.End {
		actions = append(actions, codeActionSurroundTryCatch(uri, params.Range))
	}

	if len(actions) == 0 {
		return []lsp.CodeAction{}, nil
	}
	return actions, nil
}

func codeActionRemoveCTE(uri, text string, diag lsp.Diagnostic) *lsp.CodeAction {
	cteName := extractRangeText(text,
		diag.Range.Start.Line, diag.Range.Start.Character,
		diag.Range.End.Line, diag.Range.End.Character)
	if cteName == "" {
		return nil
	}

	batchText, adjustedLine := parser.BatchAtLine(text, diag.Range.Start.Line)
	batchStartLine := diag.Range.Start.Line - adjustedLine
	parsed, err := parser.Parse(batchText)
	if err != nil {
		return nil
	}

	toks := parseutil.FlattenTokens(parsed)

	// Find WITH keyword and collect CTE spans
	type cteSpan struct {
		name     string
		nameIdx  int // index of name token
		endIdx   int // index of closing RParen token
		commaIdx int // index of comma after this CTE (-1 if none)
	}

	var withIdx int = -1
	var ctes []cteSpan

	for i := 0; i < len(toks); i++ {
		if !parseutil.MatchKeyword(toks[i], "WITH") {
			continue
		}
		// Check next non-WS token is not '(' (table hint)
		j := parseutil.SkipWS(toks, i+1)
		if j >= len(toks) || toks[j].Kind == token.LParen {
			continue
		}
		withIdx = i
		i = j
		for i < len(toks) {
			i = parseutil.SkipWS(toks, i)
			if i >= len(toks) {
				break
			}
			cs := cteSpan{nameIdx: i, commaIdx: -1}
			w, ok := toks[i].Value.(*token.SQLWord)
			if ok {
				cs.name = w.Value
			} else {
				cs.name = toks[i].String()
			}
			i++
			i = parseutil.SkipWS(toks, i)
			if i >= len(toks) || !parseutil.MatchKeyword(toks[i], "AS") {
				break
			}
			i++
			i = parseutil.SkipWS(toks, i)
			if i >= len(toks) || toks[i].Kind != token.LParen {
				break
			}
			depth := 1
			i++
			for i < len(toks) && depth > 0 {
				if toks[i].Kind == token.LParen {
					depth++
				} else if toks[i].Kind == token.RParen {
					depth--
				}
				if depth == 0 {
					cs.endIdx = i
				}
				i++
			}
			ctes = append(ctes, cs)

			j := parseutil.SkipWS(toks, i)
			if j < len(toks) && toks[j].Kind == token.Comma {
				ctes[len(ctes)-1].commaIdx = j
				i = j + 1
				continue
			}
			break
		}
		break
	}

	if withIdx < 0 || len(ctes) == 0 {
		return nil
	}

	// Find the target CTE
	targetIdx := -1
	for i, cs := range ctes {
		if strings.EqualFold(cs.name, cteName) {
			targetIdx = i
			break
		}
	}
	if targetIdx < 0 {
		return nil
	}

	var editRange lsp.Range
	if len(ctes) == 1 {
		// Single CTE: remove from WITH through closing RParen
		editRange = lsp.Range{
			Start: tokPos(toks[withIdx], batchStartLine),
			End:   tokEndPos(toks[ctes[0].endIdx], batchStartLine),
		}
	} else if targetIdx == 0 {
		// First of multiple: remove from name through comma after
		editRange = lsp.Range{
			Start: tokPos(toks[ctes[0].nameIdx], batchStartLine),
			End:   tokEndPos(toks[ctes[0].commaIdx], batchStartLine),
		}
	} else {
		// Middle or last: remove from comma before through closing RParen
		prevComma := ctes[targetIdx-1].commaIdx
		editRange = lsp.Range{
			Start: tokPos(toks[prevComma], batchStartLine),
			End:   tokEndPos(toks[ctes[targetIdx].endIdx], batchStartLine),
		}
	}

	return &lsp.CodeAction{
		Title:       fmt.Sprintf("Remove unreferenced CTE '%s'", cteName),
		Kind:        lsp.CodeActionKindQuickFix,
		Diagnostics: []lsp.Diagnostic{diag},
		Edit: &lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				uri: {
					{
						Range:   editRange,
						NewText: "",
					},
				},
			},
		},
	}
}

func tokPos(tok *ast.SQLToken, batchStartLine int) lsp.Position {
	return lsp.Position{
		Line:      tok.From.Line + batchStartLine,
		Character: tok.From.Col,
	}
}

func tokEndPos(tok *ast.SQLToken, batchStartLine int) lsp.Position {
	return lsp.Position{
		Line:      tok.To.Line + batchStartLine,
		Character: tok.To.Col,
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

func codeActionAddEndCatch(uri string, diag lsp.Diagnostic) lsp.CodeAction {
	endPos := diag.Range.End
	insertLine := endPos.Line + 1
	return lsp.CodeAction{
		Title:       "Add missing END CATCH",
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
						NewText: "END CATCH\n",
					},
				},
			},
		},
	}
}

func codeActionAddCloseParen(uri string, diag lsp.Diagnostic) lsp.CodeAction {
	endPos := diag.Range.End
	return lsp.CodeAction{
		Title:       "Add missing closing parenthesis",
		Kind:        lsp.CodeActionKindQuickFix,
		Diagnostics: []lsp.Diagnostic{diag},
		Edit: &lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				uri: {
					{
						Range: lsp.Range{
							Start: endPos,
							End:   endPos,
						},
						NewText: ")",
					},
				},
			},
		},
	}
}

func codeActionRemoveDuplicateDecl(uri, text string, diag lsp.Diagnostic) lsp.CodeAction {
	// Remove the line containing the duplicate declaration
	startLine := diag.Range.Start.Line
	return lsp.CodeAction{
		Title:       "Remove duplicate declaration",
		Kind:        lsp.CodeActionKindQuickFix,
		Diagnostics: []lsp.Diagnostic{diag},
		Edit: &lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				uri: {
					{
						Range: lsp.Range{
							Start: lsp.Position{Line: startLine, Character: diag.Range.Start.Character},
							End:   lsp.Position{Line: diag.Range.End.Line, Character: diag.Range.End.Character},
						},
						NewText: "",
					},
				},
			},
		},
	}
}

func codeActionSurroundTryCatch(uri string, r lsp.Range) lsp.CodeAction {
	return lsp.CodeAction{
		Title: "Surround with TRY/CATCH",
		Kind:  lsp.CodeActionKindRefactor,
		Edit: &lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				uri: {
					{
						Range: lsp.Range{
							Start: r.Start,
							End:   r.Start,
						},
						NewText: "BEGIN TRY\n",
					},
					{
						Range: lsp.Range{
							Start: r.End,
							End:   r.End,
						},
						NewText: "\nEND TRY\nBEGIN CATCH\nEND CATCH",
					},
				},
			},
		},
	}
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
