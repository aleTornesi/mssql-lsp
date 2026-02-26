package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/atornesi/tsql-ls/dialect"
	"github.com/atornesi/tsql-ls/pkg/completer"
	"github.com/atornesi/tsql-ls/pkg/lsp"
)

func (s *Server) handleTextDocumentCompletion(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.CompletionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	c := completer.NewCompleter(s.worker.Cache())
	completionItems, err := c.Complete(f.Text, params, s.getConfig().LowercaseKeywords)
	if err != nil {
		return nil, err
	}
	return completionItems, nil
}

func (s *Server) handleCompletionItemResolve(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var item lsp.CompletionItem
	if err := json.Unmarshal(*req.Params, &item); err != nil {
		return nil, err
	}

	// Add documentation based on item kind
	switch item.Kind {
	case lsp.FunctionCompletion:
		builtin := dialect.LookupBuiltinFunction(strings.ToUpper(item.Label))
		if builtin != nil && len(builtin.Signatures) > 0 {
			sig := builtin.Signatures[0]
			item.Documentation = &lsp.MarkupContent{
				Kind:  lsp.Markdown,
				Value: fmt.Sprintf("```\n%s\n```\n%s", sig.Label, sig.Doc),
			}
		}
	case lsp.KeywordCompletion:
		item.Documentation = &lsp.MarkupContent{
			Kind:  lsp.PlainText,
			Value: "T-SQL keyword",
		}
	case lsp.SnippetCompletion:
		if item.InsertText != "" {
			item.Documentation = &lsp.MarkupContent{
				Kind:  lsp.Markdown,
				Value: fmt.Sprintf("```sql\n%s\n```", item.InsertText),
			}
		}
	}

	return item, nil
}
