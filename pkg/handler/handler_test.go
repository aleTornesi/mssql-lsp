package handler

import (
	"context"
	"errors"
	"log"
	"net"
	"reflect"
	"testing"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/atornesi/tsql-ls/pkg/config"
	"github.com/atornesi/tsql-ls/pkg/lsp"
)

const testFileURI = "file:///Users/octref/Code/css-test/test.sql"

type TestContext struct {
	h          jsonrpc2.Handler
	conn       *jsonrpc2.Conn
	connServer *jsonrpc2.Conn
	server     *Server
	ctx        context.Context
}

func newTestContext() *TestContext {
	server := NewServer()
	handler := jsonrpc2.HandlerWithError(server.Handle)
	ctx := context.Background()
	return &TestContext{
		h:      handler,
		ctx:    ctx,
		server: server,
	}
}

func (tx *TestContext) setup(t *testing.T) {
	t.Helper()
	tx.initServer(t)
}

func (tx *TestContext) tearDown() {
	if tx.conn != nil {
		if err := tx.conn.Close(); err != nil {
			log.Fatal("conn.Close:", err)
		}
	}

	if tx.connServer != nil {
		if err := tx.connServer.Close(); err != nil {
			if !errors.Is(err, jsonrpc2.ErrClosed) {
				log.Fatal("connServer.Close:", err)
			}
		}
	}
}

func (tx *TestContext) initServer(t *testing.T) {
	t.Helper()

	// Prepare the server and client connection.
	client, server := net.Pipe()
	tx.connServer = jsonrpc2.NewConn(tx.ctx, jsonrpc2.NewBufferedStream(server, jsonrpc2.VSCodeObjectCodec{}), tx.h)
	tx.conn = jsonrpc2.NewConn(tx.ctx, jsonrpc2.NewBufferedStream(client, jsonrpc2.VSCodeObjectCodec{}), tx.h)

	// Initialize Language Server
	params := lsp.InitializeParams{
		InitializationOptions: lsp.InitializeOptions{},
	}
	if err := tx.conn.Call(tx.ctx, "initialize", params, nil); err != nil {
		t.Fatal("conn.Call initialize:", err)
	}
}

func (tx *TestContext) addWorkspaceConfig(t *testing.T, cfg *config.Config) {
	didChangeConfigurationParams := lsp.DidChangeConfigurationParams{
		Settings: struct {
			SQLS *config.Config "json:\"sqls\""
		}{
			SQLS: cfg,
		},
	}
	if err := tx.conn.Call(tx.ctx, "workspace/didChangeConfiguration", didChangeConfigurationParams, nil); err != nil {
		t.Fatal("conn.Call workspace/didChangeConfiguration:", err)
	}
}

func (tx *TestContext) textDocumentDidOpen(t *testing.T, uri, input string) {
	didOpenParams := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        testFileURI,
			LanguageID: "sql",
			Version:    0,
			Text:       input,
		},
	}
	if err := tx.conn.Call(tx.ctx, "textDocument/didOpen", didOpenParams, nil); err != nil {
		t.Fatal("conn.Call textDocument/didOpen:", err)
	}
	tx.testFile(t, didOpenParams.TextDocument.URI, didOpenParams.TextDocument.Text)
}

func TestInitialized(t *testing.T) {
	tx := newTestContext()
	tx.setup(t)
	defer tx.tearDown()

	want := lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TDSKIncremental,
			HoverProvider:    true,
			CompletionProvider: &lsp.CompletionOptions{
				ResolveProvider:   true,
				TriggerCharacters: []string{"(", "."},
			},
			SignatureHelpProvider: &lsp.SignatureHelpOptions{
				TriggerCharacters:   []string{"(", ","},
				RetriggerCharacters: []string{"(", ","},
				WorkDoneProgressOptions: lsp.WorkDoneProgressOptions{
					WorkDoneProgress: false,
				},
			},
			CodeActionProvider: map[string]interface{}{
				"CodeActionKinds": []interface{}{"quickfix", "refactor"},
			},
			DefinitionProvider:              true,
			TypeDefinitionProvider:          true,
			DeclarationProvider:             true,
			ReferencesProvider:              true,
			DocumentHighlightProvider:       true,
			DocumentFormattingProvider:      true,
			DocumentRangeFormattingProvider: true,
			DocumentOnTypeFormattingProvider: &lsp.DocumentOnTypeFormattingOptions{
				FirstTriggerCharacter: "\n",
			},
			RenameProvider: map[string]interface{}{
				"prepareProvider": true,
			},
			SelectionRangeProvider:         true,
			DocumentSymbolProvider:         true,
			WorkspaceSymbolProvider:        true,
			FoldingRangeProvider:           true,
			LinkedEditingRangeProvider:     true,
			InlayHintProvider:              true,
			SemanticTokensProvider: &lsp.SemanticTokensOptions{
				Legend: lsp.SemanticTokensLegend{
					TokenTypes:     semanticTokenTypes,
					TokenModifiers: semanticTokenModifiers,
				},
				Full:  map[string]interface{}{"delta": true},
				Range: true,
			},
		},
	}
	var got lsp.InitializeResult
	params := lsp.InitializeParams{
		InitializationOptions: lsp.InitializeOptions{},
	}
	if err := tx.conn.Call(tx.ctx, "initialize", params, &got); err != nil {
		t.Fatal("conn.Call initialize:", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("not match \n%+v\n%+v", want, got)
	}
}

func TestFileWatch(t *testing.T) {
	tx := newTestContext()
	tx.setup(t)
	defer tx.tearDown()

	uri := "file:///Users/octref/Code/css-test/test.sql"
	openText := "SELECT * FROM todo ORDER BY id ASC"
	changeText := "SELECT * FROM todo ORDER BY name ASC"

	didOpenParams := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        uri,
			LanguageID: "sql",
			Version:    0,
			Text:       openText,
		},
	}
	if err := tx.conn.Call(tx.ctx, "textDocument/didOpen", didOpenParams, nil); err != nil {
		t.Fatal("conn.Call textDocument/didOpen:", err)
	}
	tx.testFile(t, didOpenParams.TextDocument.URI, didOpenParams.TextDocument.Text)

	didChangeParams := lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			URI:     uri,
			Version: 1,
		},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{
			{
				Text: changeText,
			},
		},
	}
	if err := tx.conn.Call(tx.ctx, "textDocument/didChange", didChangeParams, nil); err != nil {
		t.Fatal("conn.Call textDocument/didChange:", err)
	}
	tx.testFile(t, didChangeParams.TextDocument.URI, changeText)

	didSaveParams := lsp.DidSaveTextDocumentParams{
		Text:         openText,
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
	}
	if err := tx.conn.Call(tx.ctx, "textDocument/didSave", didSaveParams, nil); err != nil {
		t.Fatal("conn.Call textDocument/didSave:", err)
	}
	tx.testFile(t, didSaveParams.TextDocument.URI, didSaveParams.Text)

	didCloseParams := lsp.DidCloseTextDocumentParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
	}
	if err := tx.conn.Call(tx.ctx, "textDocument/didClose", didCloseParams, nil); err != nil {
		t.Fatal("conn.Call textDocument/didClose:", err)
	}
	_, ok := tx.server.files[didCloseParams.TextDocument.URI]
	if ok {
		t.Errorf("found opened file. URI:%s", didCloseParams.TextDocument.URI)
	}
}

func TestApplyContentChange(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		change lsp.TextDocumentContentChangeEvent
		want   string
	}{
		{
			name:   "full replacement nil range",
			text:   "SELECT 1",
			change: lsp.TextDocumentContentChangeEvent{Text: "SELECT 2"},
			want:   "SELECT 2",
		},
		{
			name: "insert at beginning",
			text: "SELECT 1",
			change: lsp.TextDocumentContentChangeEvent{
				Range: &lsp.Range{
					Start: lsp.Position{Line: 0, Character: 0},
					End:   lsp.Position{Line: 0, Character: 0},
				},
				Text: "-- comment\n",
			},
			want: "-- comment\nSELECT 1",
		},
		{
			name: "replace word",
			text: "SELECT id FROM users",
			change: lsp.TextDocumentContentChangeEvent{
				Range: &lsp.Range{
					Start: lsp.Position{Line: 0, Character: 7},
					End:   lsp.Position{Line: 0, Character: 9},
				},
				Text: "name",
			},
			want: "SELECT name FROM users",
		},
		{
			name: "delete text",
			text: "SELECT id, name FROM users",
			change: lsp.TextDocumentContentChangeEvent{
				Range: &lsp.Range{
					Start: lsp.Position{Line: 0, Character: 9},
					End:   lsp.Position{Line: 0, Character: 15},
				},
				Text: "",
			},
			want: "SELECT id FROM users",
		},
		{
			name: "multiline insert",
			text: "SELECT 1\nFROM t",
			change: lsp.TextDocumentContentChangeEvent{
				Range: &lsp.Range{
					Start: lsp.Position{Line: 1, Character: 5},
					End:   lsp.Position{Line: 1, Character: 6},
				},
				Text: "users",
			},
			want: "SELECT 1\nFROM users",
		},
		{
			name: "append at EOF",
			text: "SELECT 1",
			change: lsp.TextDocumentContentChangeEvent{
				Range: &lsp.Range{
					Start: lsp.Position{Line: 0, Character: 8},
					End:   lsp.Position{Line: 0, Character: 8},
				},
				Text: "\nGO",
			},
			want: "SELECT 1\nGO",
		},
		{
			name: "utf16 surrogate pair emoji",
			text: "-- 😀 test\nSELECT 1",
			change: lsp.TextDocumentContentChangeEvent{
				Range: &lsp.Range{
					Start: lsp.Position{Line: 1, Character: 7},
					End:   lsp.Position{Line: 1, Character: 8},
				},
				Text: "2",
			},
			want: "-- 😀 test\nSELECT 2",
		},
		{
			name: "edit after emoji on same line",
			text: "-- 😀 SELECT 1",
			change: lsp.TextDocumentContentChangeEvent{
				Range: &lsp.Range{
					Start: lsp.Position{Line: 0, Character: 5}, // after "-- 😀" (😀 is 2 UTF-16 units)
					End:   lsp.Position{Line: 0, Character: 5},
				},
				Text: "!",
			},
			want: "-- 😀! SELECT 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyContentChange(tt.text, tt.change)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMultipleSequentialEdits(t *testing.T) {
	text := "SELECT 1"
	changes := []lsp.TextDocumentContentChangeEvent{
		{
			Range: &lsp.Range{
				Start: lsp.Position{Line: 0, Character: 7},
				End:   lsp.Position{Line: 0, Character: 8},
			},
			Text: "2",
		},
		{
			Range: &lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 6},
			},
			Text: "INSERT",
		},
	}
	for _, c := range changes {
		text = applyContentChange(text, c)
	}
	want := "INSERT 2"
	if text != want {
		t.Errorf("got %q, want %q", text, want)
	}
}

func (tx *TestContext) testFile(t *testing.T, uri, text string) {
	f, ok := tx.server.files[uri]
	if !ok {
		t.Errorf("not found opened file. URI:%s", uri)
	}
	if f.Text != text {
		t.Errorf("not match %s. got: %s", text, f.Text)
	}
}
