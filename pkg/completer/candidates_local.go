package completer

import (
	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
	"github.com/aleTornesi/mssql-lsp/parser/parseutil"
)

const (
	CompletionTypeVariable  completionType = iota + 100
	CompletionTypeCTE
	CompletionTypeTempTable
)

func localSymbolCandidates(st *parseutil.SymbolTable) []lsp.CompletionItem {
	var items []lsp.CompletionItem
	for _, sym := range st.Symbols {
		var kind lsp.CompletionItemKind
		var detail string
		switch sym.Kind {
		case parseutil.SymbolVariable:
			kind = lsp.VariableCompletion
			detail = sym.DataType
		case parseutil.SymbolCTE:
			kind = lsp.ClassCompletion
			detail = "CTE"
		case parseutil.SymbolTempTable:
			kind = lsp.ClassCompletion
			detail = "Temp Table"
		}
		items = append(items, lsp.CompletionItem{
			Label:  sym.Name,
			Kind:   kind,
			Detail: detail,
		})
	}
	return items
}
