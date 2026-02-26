package handler

import (
	"context"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/aleTornesi/mssql-lsp/diagnostic"
	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
)

type PublishDiagnosticsParams struct {
	URI         string           `json:"uri"`
	Diagnostics []lsp.Diagnostic `json:"diagnostics"`
}

func (s *Server) publishDiagnostics(ctx context.Context, conn *jsonrpc2.Conn, uri string, text string) {
	diags := diagnostic.Analyze(text)

	lspDiags := make([]lsp.Diagnostic, 0, len(diags))
	source := "tsql-ls"
	for _, d := range diags {
		code := d.Code
		lspDiags = append(lspDiags, lsp.Diagnostic{
			Range: lsp.Range{
				Start: lsp.Position{Line: d.From.Line, Character: d.From.Col},
				End:   lsp.Position{Line: d.To.Line, Character: d.To.Col},
			},
			Severity: int(d.Severity),
			Code:     &code,
			Source:   &source,
			Message:  d.Message,
		})
	}

	conn.Notify(ctx, "textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: lspDiags,
	})
}

func (s *Server) clearDiagnostics(ctx context.Context, conn *jsonrpc2.Conn, uri string) {
	conn.Notify(ctx, "textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: []lsp.Diagnostic{},
	})
}
