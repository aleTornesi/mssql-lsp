package diagnostic

import (
	"fmt"
	"strings"

	"github.com/atornesi/tsql-ls/ast"
	"github.com/atornesi/tsql-ls/parser/parseutil"
)

// CheckSemantics runs semantic checks on a parsed batch.
func CheckSemantics(parsed ast.TokenList, c *Collector) {
	st := parseutil.ExtractSymbols(parsed)
	checkDuplicateVariables(st, c)
	checkUnreferencedCTEs(parsed, st, c)
}

func checkDuplicateVariables(st *parseutil.SymbolTable, c *Collector) {
	seen := make(map[string]*parseutil.Symbol)
	for _, sym := range st.Symbols {
		if sym.Kind != parseutil.SymbolVariable {
			continue
		}
		key := strings.ToUpper(sym.Name)
		if _, exists := seen[key]; exists {
			c.Add(Diagnostic{
				From:     sym.Pos,
				To:       sym.EndPos,
				Severity: Warning,
				Message:  fmt.Sprintf("duplicate variable declaration: '%s'", sym.Name),
				Code:     CodeDuplicateVariable,
			})
		} else {
			seen[key] = sym
		}
	}
}

func checkUnreferencedCTEs(parsed ast.TokenList, st *parseutil.SymbolTable, c *Collector) {
	for _, sym := range st.Symbols {
		if sym.Kind != parseutil.SymbolCTE {
			continue
		}
		refs := st.FindReferences(parsed, sym.Name)
		// Filter out the definition position itself
		usageCount := 0
		for _, ref := range refs {
			if ref.Line != sym.Pos.Line || ref.Col != sym.Pos.Col {
				usageCount++
			}
		}
		if usageCount == 0 {
			c.Add(Diagnostic{
				From:     sym.Pos,
				To:       sym.EndPos,
				Severity: Warning,
				Message:  fmt.Sprintf("unreferenced CTE: '%s'", sym.Name),
				Code:     CodeUnreferencedCTE,
			})
		}
	}
}
