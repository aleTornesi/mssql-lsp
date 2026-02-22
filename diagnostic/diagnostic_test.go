package diagnostic

import (
	"testing"
)

func TestAnalyze_CleanFile(t *testing.T) {
	diags := Analyze("SELECT 1")
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d: %+v", len(diags), diags)
	}
}

func TestAnalyze_UnclosedMultilineComment(t *testing.T) {
	diags := Analyze("SELECT /* unclosed comment")
	assertHasCode(t, diags, CodeUnclosedComment)
}

func TestAnalyze_UnclosedString(t *testing.T) {
	diags := Analyze("SELECT 'unclosed")
	assertHasCode(t, diags, CodeUnclosedString)
}

func TestAnalyze_UnclosedBracketIdentifier(t *testing.T) {
	diags := Analyze("SELECT [unclosed")
	assertHasCode(t, diags, CodeUnclosedBracket)
}

func TestAnalyze_UnclosedBegin(t *testing.T) {
	diags := Analyze("BEGIN\n  SELECT 1")
	assertHasCode(t, diags, CodeUnclosedBegin)
}

func TestAnalyze_UnclosedTryCatch(t *testing.T) {
	diags := Analyze("BEGIN TRY\n  SELECT 1")
	assertHasCode(t, diags, CodeUnclosedTryCatch)
}

func TestAnalyze_UnclosedParenthesis(t *testing.T) {
	diags := Analyze("SELECT (1 + 2")
	assertHasCode(t, diags, CodeUnclosedParen)
}

func TestAnalyze_UnclosedCase(t *testing.T) {
	diags := Analyze("SELECT CASE WHEN 1=1 THEN 'a'")
	assertHasCode(t, diags, CodeUnclosedCase)
}

func TestAnalyze_MultiBatch_Offsets(t *testing.T) {
	// Second batch has unclosed string, should be offset by GO line
	input := "SELECT 1\nGO\nSELECT 'unclosed"
	diags := Analyze(input)
	assertHasCode(t, diags, CodeUnclosedString)
	for _, d := range diags {
		if d.Code == CodeUnclosedString {
			if d.From.Line != 2 {
				t.Errorf("expected diagnostic on line 2, got line %d", d.From.Line)
			}
		}
	}
}

func TestAnalyze_MultipleErrors(t *testing.T) {
	// Use separate batches so each error is independent
	input := "SELECT /* unclosed\nGO\nSELECT 'unclosed"
	diags := Analyze(input)
	codes := map[string]bool{}
	for _, d := range diags {
		codes[d.Code] = true
	}
	if !codes[CodeUnclosedString] {
		t.Error("expected TSQL002")
	}
	if !codes[CodeUnclosedComment] {
		t.Error("expected TSQL001")
	}
}

func TestAnalyze_ClosedConstructs(t *testing.T) {
	tests := []struct {
		name string
		sql  string
	}{
		{"closed comment", "SELECT /* comment */ 1"},
		{"closed string", "SELECT 'hello'"},
		{"closed bracket", "SELECT [col]"},
		{"closed begin", "BEGIN\n  SELECT 1\nEND"},
		{"closed try/catch", "BEGIN TRY\n  SELECT 1\nEND TRY\nBEGIN CATCH\n  SELECT 2\nEND CATCH"},
		{"closed parens", "SELECT (1 + 2)"},
		{"closed case", "SELECT CASE WHEN 1=1 THEN 'a' END"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := Analyze(tt.sql)
			if len(diags) != 0 {
				t.Errorf("expected 0 diagnostics, got %d: %+v", len(diags), diags)
			}
		})
	}
}

func assertHasCode(t *testing.T, diags []Diagnostic, code string) {
	t.Helper()
	for _, d := range diags {
		if d.Code == code {
			return
		}
	}
	codes := make([]string, len(diags))
	for i, d := range diags {
		codes[i] = d.Code
	}
	t.Errorf("expected diagnostic code %s, got codes: %v", code, codes)
}
