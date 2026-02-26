package completer

import (
	"strings"
	"testing"

	"github.com/atornesi/tsql-ls/pkg/lsp"
)

func TestComplete_ExecSystemProc(t *testing.T) {
	c := NewCompleter(nil)
	items, err := c.Complete("EXEC sp_", lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			Position: lsp.Position{Line: 0, Character: 8},
		},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range items {
		if item.Label == "sp_help" {
			found = true
			if item.Kind != lsp.FunctionCompletion {
				t.Errorf("sp_help Kind = %d, want FunctionCompletion", item.Kind)
			}
			break
		}
	}
	if !found {
		t.Error("expected sp_help in EXEC completions")
	}
}

func TestComplete_ExecPrefixFilter(t *testing.T) {
	c := NewCompleter(nil)
	items, err := c.Complete("EXEC xp_", lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			Position: lsp.Position{Line: 0, Character: 8},
		},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if !strings.HasPrefix(strings.ToUpper(item.Label), "XP_") {
			t.Errorf("unexpected item %q with prefix xp_", item.Label)
		}
	}
	if len(items) == 0 {
		t.Error("expected xp_ completions")
	}
}

func TestComplete_FromSysViews(t *testing.T) {
	c := NewCompleter(nil)
	items, err := c.Complete("SELECT * FROM sys.", lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			Position: lsp.Position{Line: 0, Character: 18},
		},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range items {
		if item.Label == "sys.tables" || item.Label == "sys.objects" {
			found = true
			break
		}
	}
	if !found {
		t.Logf("got %d items", len(items))
		for _, item := range items {
			t.Logf("  %q (%d)", item.Label, item.Kind)
		}
		t.Error("expected sys.tables or sys.objects in FROM sys. completions")
	}
}
