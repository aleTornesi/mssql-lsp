package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
)

func (s *Server) handleTextDocumentOnTypeFormatting(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DocumentOnTypeFormattingParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	edits := onTypeFormatting(f.Text, params)
	if len(edits) == 0 {
		return nil, nil
	}
	return edits, nil
}

func onTypeFormatting(text string, params lsp.DocumentOnTypeFormattingParams) []lsp.TextEdit {
	if params.Ch != "\n" {
		return nil
	}

	// The cursor is now on the new line (params.Position).
	// Look at the previous line to decide indentation.
	curLine := params.Position.Line
	if curLine == 0 {
		return nil
	}

	lines := strings.Split(text, "\n")
	if curLine-1 >= len(lines) {
		return nil
	}

	prevLine := lines[curLine-1]
	prevTrimmed := strings.TrimSpace(prevLine)
	prevUpper := strings.ToUpper(prevTrimmed)

	// Calculate current indentation of previous line
	prevIndent := leadingWhitespace(prevLine)

	indent := indentString(params.Options)

	// If previous line ends with BEGIN, AS, THEN, ELSE, or starts a block, add indent
	if endsWithBlockOpener(prevUpper) {
		newIndent := prevIndent + indent
		return []lsp.TextEdit{
			{
				Range: lsp.Range{
					Start: lsp.Position{Line: curLine, Character: 0},
					End:   lsp.Position{Line: curLine, Character: 0},
				},
				NewText: newIndent,
			},
		}
	}

	// If current line (the new line, which might have text after it from splitting)
	// starts with END, END CATCH, dedent
	if curLine < len(lines) {
		curTrimmed := strings.TrimSpace(lines[curLine])
		curUpper := strings.ToUpper(curTrimmed)
		if startsWithBlockCloser(curUpper) {
			newIndent := dedent(prevIndent, indent)
			return []lsp.TextEdit{
				{
					Range: lsp.Range{
						Start: lsp.Position{Line: curLine, Character: 0},
						End:   lsp.Position{Line: curLine, Character: 0},
					},
					NewText: newIndent,
				},
			}
		}
	}

	return nil
}

func endsWithBlockOpener(upper string) bool {
	openers := []string{"BEGIN", "BEGIN TRY", "BEGIN CATCH", "THEN", "ELSE", "AS"}
	for _, opener := range openers {
		if strings.HasSuffix(upper, opener) {
			return true
		}
	}
	return false
}

func startsWithBlockCloser(upper string) bool {
	closers := []string{"END", "END TRY", "END CATCH"}
	for _, closer := range closers {
		if upper == closer || strings.HasPrefix(upper, closer+" ") || strings.HasPrefix(upper, closer+";") {
			return true
		}
	}
	return false
}

func leadingWhitespace(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	return line[:len(line)-len(trimmed)]
}

func indentString(opts lsp.FormattingOptions) string {
	if opts.InsertSpaces {
		size := int(opts.TabSize)
		if size <= 0 {
			size = 4
		}
		return strings.Repeat(" ", size)
	}
	return "\t"
}

func dedent(current, indent string) string {
	if strings.HasSuffix(current, indent) {
		return current[:len(current)-len(indent)]
	}
	return current
}
