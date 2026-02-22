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
	checkSetUndeclaredVariable(parsed, st, c)
	checkTransactionBalance(parsed, c)
	checkDuplicateCTENames(st, c)
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

func checkSetUndeclaredVariable(parsed ast.TokenList, st *parseutil.SymbolTable, c *Collector) {
	toks := flattenAllTokens(parsed)
	for i := 0; i < len(toks); i++ {
		if !matchKW(toks[i], "SET") {
			continue
		}
		j := i + 1
		for j < len(toks) && isWSToken(toks[j]) {
			j++
		}
		if j >= len(toks) {
			continue
		}
		name := toks[j].String()
		if !strings.HasPrefix(name, "@") {
			continue
		}
		if strings.HasPrefix(name, "@@") {
			if builtinVariables[strings.ToUpper(name)] {
				continue
			}
		}
		if st.Lookup(name) != nil {
			continue
		}
		c.Add(Diagnostic{
			From:     toks[j].Pos(),
			To:       toks[j].End(),
			Severity: Warning,
			Message:  fmt.Sprintf("SET on undeclared variable: '%s'", name),
			Code:     CodeSetUndeclaredVariable,
		})
	}
}

func checkTransactionBalance(parsed ast.TokenList, c *Collector) {
	toks := flattenAllTokens(parsed)
	beginCount := 0
	commitCount := 0
	rollbackCount := 0
	var firstBegin token.Pos
	for i := 0; i < len(toks); i++ {
		if matchKW(toks[i], "BEGIN") {
			j := i + 1
			for j < len(toks) && isWSToken(toks[j]) {
				j++
			}
			if j < len(toks) && (matchKW(toks[j], "TRANSACTION") || matchKW(toks[j], "TRAN")) {
				if beginCount == 0 {
					firstBegin = toks[i].Pos()
				}
				beginCount++
			}
		}
		if matchKW(toks[i], "COMMIT") {
			j := i + 1
			for j < len(toks) && isWSToken(toks[j]) {
				j++
			}
			if j >= len(toks) || matchKW(toks[j], "TRANSACTION") || matchKW(toks[j], "TRAN") || toks[j].String() == ";" {
				commitCount++
			}
		}
		if matchKW(toks[i], "ROLLBACK") {
			j := i + 1
			for j < len(toks) && isWSToken(toks[j]) {
				j++
			}
			if j >= len(toks) || matchKW(toks[j], "TRANSACTION") || matchKW(toks[j], "TRAN") || toks[j].String() == ";" {
				rollbackCount++
			}
		}
	}
	endCount := commitCount + rollbackCount
	if beginCount > 0 && endCount == 0 {
		c.Add(Diagnostic{
			From:     firstBegin,
			To:       token.Pos{Line: firstBegin.Line, Col: firstBegin.Col + 5},
			Severity: Warning,
			Message:  fmt.Sprintf("BEGIN TRANSACTION without matching COMMIT/ROLLBACK (%d open)", beginCount),
			Code:     CodeTransactionMismatch,
		})
	}
}

func checkDuplicateCTENames(st *parseutil.SymbolTable, c *Collector) {
	seen := make(map[string]*parseutil.Symbol)
	for _, sym := range st.Symbols {
		if sym.Kind != parseutil.SymbolCTE {
			continue
		}
		key := strings.ToUpper(sym.Name)
		if _, exists := seen[key]; exists {
			c.Add(Diagnostic{
				From:     sym.Pos,
				To:       sym.EndPos,
				Severity: Error,
				Message:  fmt.Sprintf("duplicate CTE name: '%s'", sym.Name),
				Code:     CodeDuplicateCTE,
			})
		} else {
			seen[key] = sym
		}
	}
}

// flattenAllTokens collects all leaf nodes preserving order.
func flattenAllTokens(tl ast.TokenList) []ast.Node {
	var result []ast.Node
	for _, node := range tl.GetTokens() {
		if inner, ok := node.(ast.TokenList); ok {
			result = append(result, flattenAllTokens(inner)...)
		} else {
			result = append(result, node)
		}
	}
	return result
}

func matchKW(node ast.Node, kw string) bool {
	return strings.EqualFold(strings.TrimSpace(node.String()), kw)
}

func isWSToken(node ast.Node) bool {
	s := node.String()
	return strings.TrimSpace(s) == ""
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
