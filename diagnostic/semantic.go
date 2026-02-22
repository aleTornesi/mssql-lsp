package diagnostic

import (
	"fmt"
	"strings"

	"github.com/atornesi/tsql-ls/ast"
	"github.com/atornesi/tsql-ls/parser/parseutil"
	"github.com/atornesi/tsql-ls/token"
)

// CheckSemantics runs semantic checks on a parsed batch.
func CheckSemantics(parsed ast.TokenList, c *Collector) {
	st := parseutil.ExtractSymbols(parsed)
	checkDuplicateVariables(st, c)
	checkUnreferencedCTEs(parsed, st, c)
	checkUndefinedVariables(parsed, st, c)
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

// builtinVariables are system @@ variables that don't need DECLARE.
var builtinVariables = map[string]bool{
	"@@ROWCOUNT":       true,
	"@@ERROR":          true,
	"@@IDENTITY":       true,
	"@@TRANCOUNT":      true,
	"@@FETCH_STATUS":   true,
	"@@SPID":           true,
	"@@VERSION":        true,
	"@@SERVERNAME":     true,
	"@@SERVICENAME":    true,
	"@@NESTLEVEL":      true,
	"@@PROCID":         true,
	"@@DBTS":           true,
	"@@LANGID":         true,
	"@@LANGUAGE":       true,
	"@@LOCK_TIMEOUT":   true,
	"@@MAX_CONNECTIONS": true,
	"@@MAX_PRECISION":  true,
	"@@OPTIONS":        true,
	"@@REMSERVER":      true,
	"@@CURSOR_ROWS":    true,
	"@@DATEFIRST":      true,
	"@@CONNECTIONS":    true,
	"@@CPU_BUSY":       true,
	"@@IDLE":           true,
	"@@IO_BUSY":        true,
	"@@PACKET_ERRORS":  true,
	"@@PACK_RECEIVED":  true,
	"@@PACK_SENT":      true,
	"@@TIMETICKS":      true,
	"@@TOTAL_ERRORS":   true,
	"@@TOTAL_READ":     true,
	"@@TOTAL_WRITE":    true,
	"@@SCOPE_IDENTITY": true,
}

func checkUndefinedVariables(parsed ast.TokenList, st *parseutil.SymbolTable, c *Collector) {
	collectUndefinedVars(parsed, st, c)
}

func collectUndefinedVars(tl ast.TokenList, st *parseutil.SymbolTable, c *Collector) {
	for _, node := range tl.GetTokens() {
		switch n := node.(type) {
		case *ast.Identifier:
			checkVarToken(n.String(), n.Pos(), n.End(), st, c)
		case *ast.Item:
			if n.Tok != nil {
				checkVarToken(n.Tok.String(), n.Tok.From, n.Tok.To, st, c)
			}
		case ast.TokenList:
			collectUndefinedVars(n, st, c)
		}
	}
}

func checkVarToken(name string, from, to token.Pos, st *parseutil.SymbolTable, c *Collector) {
	if !strings.HasPrefix(name, "@") {
		return
	}
	// Skip built-in @@ variables
	if strings.HasPrefix(name, "@@") {
		if builtinVariables[strings.ToUpper(name)] {
			return
		}
	}
	// Check if declared in symbol table
	if st.Lookup(name) != nil {
		return
	}
	c.Add(Diagnostic{
		From:     from,
		To:       to,
		Severity: Warning,
		Message:  fmt.Sprintf("undefined variable: '%s'", name),
		Code:     CodeUndefinedVariable,
	})
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
