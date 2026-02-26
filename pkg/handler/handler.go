package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"unicode/utf8"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/aleTornesi/mssql-lsp/pkg/config"
	"github.com/aleTornesi/mssql-lsp/pkg/database"
	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
)

type Server struct {
	SpecificFileCfg *config.Config
	DefaultFileCfg  *config.Config
	WSCfg           *config.Config

	worker             *database.Worker
	files              map[string]*File
	semanticTokenCache map[string]*cachedSemanticTokens
}

type cachedSemanticTokens struct {
	resultID string
	data     []uint32
}

type File struct {
	LanguageID string
	Text       string
}

func NewServer() *Server {
	worker := database.NewWorker()
	worker.Start()

	return &Server{
		files:              make(map[string]*File),
		worker:             worker,
		semanticTokenCache: make(map[string]*cachedSemanticTokens),
	}
}

func panicf(r interface{}, format string, v ...interface{}) error {
	if r != nil {
		const size = 64 << 10
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		id := fmt.Sprintf(format, v...)
		log.Printf("panic serving %s: %v\n%s", id, r, string(buf))
		return fmt.Errorf("unexpected panic: %v", r)
	}
	return nil
}

func (s *Server) Stop() error {
	s.worker.Stop()
	return nil
}

func (s *Server) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	defer func() {
		if perr := panicf(recover(), "%v", req.Method); perr != nil {
			err = perr
		}
	}()
	res, err := s.handle(ctx, conn, req)
	if err != nil {
		log.Printf("error serving, %+v\n", err)
	}
	return res, err
}

func (s *Server) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(ctx, conn, req)
	case "initialized":
		return
	case "shutdown":
		return s.handleShutdown(ctx, conn, req)
	case "exit":
		return s.handleExit(ctx, conn, req)
	case "textDocument/didOpen":
		return s.handleTextDocumentDidOpen(ctx, conn, req)
	case "textDocument/didChange":
		return s.handleTextDocumentDidChange(ctx, conn, req)
	case "textDocument/didSave":
		return s.handleTextDocumentDidSave(ctx, conn, req)
	case "textDocument/didClose":
		return s.handleTextDocumentDidClose(ctx, conn, req)
	case "textDocument/completion":
		return s.handleTextDocumentCompletion(ctx, conn, req)
	case "completionItem/resolve":
		return s.handleCompletionItemResolve(ctx, conn, req)
	case "textDocument/hover":
		return s.handleTextDocumentHover(ctx, conn, req)
	case "textDocument/codeAction":
		return s.handleTextDocumentCodeAction(ctx, conn, req)
	case "workspace/didChangeConfiguration":
		return s.handleWorkspaceDidChangeConfiguration(ctx, conn, req)
	case "textDocument/formatting":
		return s.handleTextDocumentFormatting(ctx, conn, req)
	case "textDocument/rangeFormatting":
		return s.handleTextDocumentRangeFormatting(ctx, conn, req)
	case "textDocument/signatureHelp":
		return s.handleTextDocumentSignatureHelp(ctx, conn, req)
	case "textDocument/rename":
		return s.handleTextDocumentRename(ctx, conn, req)
	case "textDocument/definition":
		return s.handleDefinition(ctx, conn, req)
	case "textDocument/typeDefinition":
		return s.handleTypeDefinition(ctx, conn, req)
	case "textDocument/declaration":
		return s.handleDefinition(ctx, conn, req)
	case "textDocument/foldingRange":
		return s.handleTextDocumentFoldingRange(ctx, conn, req)
	case "textDocument/documentSymbol":
		return s.handleTextDocumentDocumentSymbol(ctx, conn, req)
	case "textDocument/references":
		return s.handleTextDocumentReferences(ctx, conn, req)
	case "textDocument/documentHighlight":
		return s.handleTextDocumentDocumentHighlight(ctx, conn, req)
	case "workspace/symbol":
		return s.handleWorkspaceSymbol(ctx, conn, req)
	case "textDocument/semanticTokens/full":
		return s.handleTextDocumentSemanticTokensFull(ctx, conn, req)
	case "textDocument/semanticTokens/range":
		return s.handleTextDocumentSemanticTokensRange(ctx, conn, req)
	case "textDocument/semanticTokens/full/delta":
		return s.handleTextDocumentSemanticTokensDelta(ctx, conn, req)
	case "textDocument/linkedEditingRange":
		return s.handleLinkedEditingRange(ctx, conn, req)
	case "textDocument/inlayHint":
		return s.handleInlayHint(ctx, conn, req)
	case "textDocument/prepareRename":
		return s.handleTextDocumentPrepareRename(ctx, conn, req)
	case "textDocument/selectionRange":
		return s.handleTextDocumentSelectionRange(ctx, conn, req)
	case "textDocument/onTypeFormatting":
		return s.handleTextDocumentOnTypeFormatting(ctx, conn, req)
	case "window/showMessage":
		return
	}
	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (s *Server) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	go s.connectDB(ctx)

	result = lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync:   lsp.TDSKIncremental,
			HoverProvider:      true,
			CodeActionProvider: lsp.CodeActionOptions{
				CodeActionKinds: []lsp.CodeActionKind{lsp.CodeActionKindQuickFix, lsp.CodeActionKindRefactor},
			},
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
			RenameProvider:                 lsp.RenameOptions{PrepareProvider: true},
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
				Full:  lsp.SemanticTokensFullOptions{Delta: true},
				Range: true,
			},
		},
	}

	return result, nil
}

func (s *Server) handleShutdown(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	return nil, nil
}

func (s *Server) handleExit(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	err = s.Stop()
	return nil, err
}

func (s *Server) handleTextDocumentDidOpen(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	if err := s.openFile(params.TextDocument.URI, params.TextDocument.LanguageID); err != nil {
		return nil, err
	}
	if err := s.updateFile(params.TextDocument.URI, params.TextDocument.Text); err != nil {
		return nil, err
	}
	s.publishDiagnostics(ctx, conn, params.TextDocument.URI, params.TextDocument.Text)
	return nil, nil
}

func (s *Server) handleTextDocumentDidChange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %v", params.TextDocument.URI)
	}

	for _, change := range params.ContentChanges {
		f.Text = applyContentChange(f.Text, change)
	}

	s.publishDiagnostics(ctx, conn, params.TextDocument.URI, f.Text)
	return nil, nil
}

func applyContentChange(text string, change lsp.TextDocumentContentChangeEvent) string {
	if change.Range == nil {
		return change.Text
	}

	startOffset := utf16PositionToByteOffset(text, change.Range.Start)
	endOffset := utf16PositionToByteOffset(text, change.Range.End)

	return text[:startOffset] + change.Text + text[endOffset:]
}

func utf16PositionToByteOffset(text string, pos lsp.Position) int {
	line := 0
	offset := 0

	// Advance to the target line.
	for offset < len(text) && line < pos.Line {
		if text[offset] == '\n' {
			line++
		}
		offset++
	}

	// Advance by UTF-16 code units within the line.
	utf16Col := 0
	for offset < len(text) && utf16Col < pos.Character {
		r, size := utf8.DecodeRuneInString(text[offset:])
		if r >= 0x10000 {
			utf16Col += 2 // surrogate pair
		} else {
			utf16Col++
		}
		offset += size
	}

	return offset
}

func (s *Server) handleTextDocumentDidSave(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DidSaveTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	if params.Text != "" {
		err = s.updateFile(params.TextDocument.URI, params.Text)
	} else {
		err = s.saveFile(params.TextDocument.URI)
	}
	if err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if ok {
		s.publishDiagnostics(ctx, conn, params.TextDocument.URI, f.Text)
	}
	return nil, nil
}

func (s *Server) handleTextDocumentDidClose(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DidCloseTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	if err := s.closeFile(params.TextDocument.URI); err != nil {
		return nil, err
	}
	s.clearDiagnostics(ctx, conn, params.TextDocument.URI)
	return nil, nil
}

func (s *Server) openFile(uri string, languageID string) error {
	f := &File{
		Text:       "",
		LanguageID: languageID,
	}
	s.files[uri] = f
	return nil
}

func (s *Server) closeFile(uri string) error {
	delete(s.files, uri)
	return nil
}

func (s *Server) updateFile(uri string, text string) error {
	f, ok := s.files[uri]
	if !ok {
		return fmt.Errorf("document not found: %v", uri)
	}
	f.Text = text
	return nil
}

func (s *Server) saveFile(uri string) error {
	return nil
}

func (s *Server) handleWorkspaceDidChangeConfiguration(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params lsp.DidChangeConfigurationParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	s.WSCfg = params.Settings.SQLS
	go s.connectDB(ctx)
	return nil, nil
}

func (s *Server) connectDB(ctx context.Context) {
	cfg := s.getConfig()
	if cfg == nil || len(cfg.Connections) == 0 {
		return
	}

	repo, err := database.Open(ctx, cfg.Connections[0])
	if err != nil {
		log.Printf("connectDB: %v", err)
		return
	}

	if err := s.worker.ReCache(ctx, repo); err != nil {
		log.Printf("connectDB cache: %v", err)
		return
	}
	log.Println("connectDB: connected successfully")
}

func (s *Server) getConfig() *config.Config {
	switch {
	case s.SpecificFileCfg != nil:
		return s.SpecificFileCfg
	case s.WSCfg != nil:
		return s.WSCfg
	case s.DefaultFileCfg != nil:
		return s.DefaultFileCfg
	default:
		return config.NewConfig()
	}
}
