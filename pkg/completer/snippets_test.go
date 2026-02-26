package completer

import (
	"strings"
	"testing"

	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
)

func TestSnippetCandidates(t *testing.T) {
	items := snippetCandidates()
	if len(items) == 0 {
		t.Fatal("expected snippet candidates")
	}

	for _, item := range items {
		if item.Kind != lsp.SnippetCompletion {
			t.Errorf("snippet %q Kind = %d, want SnippetCompletion", item.Label, item.Kind)
		}
		if item.InsertTextFormat != lsp.SnippetTextFormat {
			t.Errorf("snippet %q InsertTextFormat = %d, want SnippetTextFormat", item.Label, item.InsertTextFormat)
		}
		if !strings.Contains(item.InsertText, "$") {
			t.Errorf("snippet %q missing tabstop markers", item.Label)
		}
	}
}

func TestComplete_SnippetsAtStatementStart(t *testing.T) {
	c := NewCompleter(nil)
	items, err := c.Complete("BEG", lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			Position: lsp.Position{Line: 0, Character: 3},
		},
	}, false)
	if err != nil {
		t.Fatal(err)
	}

	foundSnippet := false
	for _, item := range items {
		if item.Kind == lsp.SnippetCompletion && strings.HasPrefix(item.Label, "BEGIN") {
			foundSnippet = true
			break
		}
	}
	if !foundSnippet {
		t.Error("expected BEGIN snippet at statement start")
	}
}

func TestComplete_NoSnippetsInSelectColumnList(t *testing.T) {
	c := NewCompleter(nil)
	items, err := c.Complete("SELECT ", lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			Position: lsp.Position{Line: 0, Character: 7},
		},
	}, false)
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if item.Kind == lsp.SnippetCompletion {
			t.Errorf("unexpected snippet %q in SELECT column list", item.Label)
		}
	}
}
