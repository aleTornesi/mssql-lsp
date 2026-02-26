package parseutil

import (
	"strings"

	"github.com/aleTornesi/mssql-lsp/ast"
	"github.com/aleTornesi/mssql-lsp/token"
)

// SymbolKind classifies a batch-scoped symbol.
type SymbolKind int

const (
	SymbolVariable  SymbolKind = iota // @variable
	SymbolCTE                         // WITH name AS (...)
	SymbolTempTable                   // #temp or ##temp
)

// Symbol represents a named declaration in a batch.
type Symbol struct {
	Name     string
	DataType string
	Kind     SymbolKind
	Pos      token.Pos // start of the symbol name token
	EndPos   token.Pos // end of the symbol name token
}

// SymbolTable holds all symbols extracted from a batch.
type SymbolTable struct {
	Symbols []*Symbol
}

// Lookup finds a symbol by name (case-insensitive for variables).
func (st *SymbolTable) Lookup(name string) *Symbol {
	upper := strings.ToUpper(name)
	for _, s := range st.Symbols {
		if strings.ToUpper(s.Name) == upper {
			return s
		}
	}
	return nil
}

// FindReferences finds all positions where name appears as an identifier in the token list.
func (st *SymbolTable) FindReferences(parsed ast.TokenList, name string) []token.Pos {
	upper := strings.ToUpper(name)
	var refs []token.Pos
	findRefsRecursive(parsed, upper, &refs)
	return refs
}

func findRefsRecursive(tl ast.TokenList, upper string, refs *[]token.Pos) {
	for _, node := range tl.GetTokens() {
		switch n := node.(type) {
		case *ast.Identifier:
			if strings.ToUpper(n.String()) == upper {
				*refs = append(*refs, n.Pos())
			}
		case *ast.MemberIdentifier:
			if strings.ToUpper(n.Child.String()) == upper {
				*refs = append(*refs, n.Child.Pos())
			}
			if strings.ToUpper(n.Parent.String()) == upper {
				*refs = append(*refs, n.Parent.Pos())
			}
		case ast.TokenList:
			findRefsRecursive(n, upper, refs)
		default:
			if item, ok := n.(*ast.Item); ok {
				if item.Tok != nil && strings.ToUpper(item.Tok.String()) == upper {
					*refs = append(*refs, item.Tok.From)
				}
			}
		}
	}
}

// ExtractSymbols walks a parsed batch and extracts all variable, CTE, and temp table declarations.
func ExtractSymbols(parsed ast.TokenList) *SymbolTable {
	st := &SymbolTable{}
	toks := flattenTokens(parsed)
	extractDeclareVariables(toks, st)
	extractCTEs(toks, st)
	extractTempTables(toks, st)
	return st
}

// FlattenTokens collects all leaf SQLTokens in order.
func FlattenTokens(tl ast.TokenList) []*ast.SQLToken {
	return flattenTokens(tl)
}

func flattenTokens(tl ast.TokenList) []*ast.SQLToken {
	var result []*ast.SQLToken
	flattenRecursive(tl, &result)
	return result
}

func flattenRecursive(tl ast.TokenList, result *[]*ast.SQLToken) {
	for _, node := range tl.GetTokens() {
		switch n := node.(type) {
		case *ast.Item:
			if n.Tok != nil {
				*result = append(*result, n.Tok)
			}
		case *ast.Identifier:
			if n.Tok != nil {
				*result = append(*result, n.Tok)
			}
		case ast.TokenList:
			flattenRecursive(n, result)
		}
	}
}

func isWhitespaceOrComment(tok *ast.SQLToken) bool {
	return tok.Kind == token.Whitespace || tok.Kind == token.Comment || tok.Kind == token.MultilineComment
}

// SkipWS advances past whitespace and comment tokens.
func SkipWS(toks []*ast.SQLToken, i int) int {
	return skipWS(toks, i)
}

func skipWS(toks []*ast.SQLToken, i int) int {
	for i < len(toks) && isWhitespaceOrComment(toks[i]) {
		i++
	}
	return i
}

// MatchKeyword checks if a token is a SQL keyword matching kw (case-insensitive).
func MatchKeyword(tok *ast.SQLToken, kw string) bool {
	return matchKeyword(tok, kw)
}

func matchKeyword(tok *ast.SQLToken, kw string) bool {
	if tok.Kind != token.SQLKeyword {
		return false
	}
	w, ok := tok.Value.(*token.SQLWord)
	if !ok {
		return false
	}
	return strings.EqualFold(w.Keyword, kw)
}

func tokString(tok *ast.SQLToken) string {
	if tok.Kind == token.SQLKeyword {
		w, ok := tok.Value.(*token.SQLWord)
		if ok {
			return w.Value
		}
	}
	return tok.String()
}

// extractDeclareVariables handles: DECLARE @x INT, @y VARCHAR(50)
func extractDeclareVariables(toks []*ast.SQLToken, st *SymbolTable) {
	for i := 0; i < len(toks); i++ {
		if !matchKeyword(toks[i], "DECLARE") {
			continue
		}
		i++
		for i < len(toks) {
			i = skipWS(toks, i)
			if i >= len(toks) {
				break
			}
			// Expect @variable
			name := tokString(toks[i])
			if !strings.HasPrefix(name, "@") {
				break
			}
			sym := &Symbol{
				Name:   name,
				Kind:   SymbolVariable,
				Pos:    toks[i].From,
				EndPos: toks[i].To,
			}
			i++
			i = skipWS(toks, i)

			// Collect data type tokens until comma, semicolon, or next statement keyword
			var dtParts []string
			for i < len(toks) {
				if toks[i].Kind == token.Comma || toks[i].Kind == token.Semicolon {
					break
				}
				if isWhitespaceOrComment(toks[i]) {
					i++
					continue
				}
				// Stop at assignment
				if toks[i].Kind == token.Eq {
					// skip past = and the value expression up to comma/semicolon
					i++
					i = skipPastExpression(toks, i)
					break
				}
				// Stop at another DECLARE or DML keyword
				if matchKeyword(toks[i], "DECLARE") || matchKeyword(toks[i], "SELECT") ||
					matchKeyword(toks[i], "SET") || matchKeyword(toks[i], "INSERT") ||
					matchKeyword(toks[i], "UPDATE") || matchKeyword(toks[i], "DELETE") ||
					matchKeyword(toks[i], "IF") || matchKeyword(toks[i], "WHILE") ||
					matchKeyword(toks[i], "BEGIN") || matchKeyword(toks[i], "EXEC") ||
					matchKeyword(toks[i], "EXECUTE") || matchKeyword(toks[i], "RETURN") ||
					matchKeyword(toks[i], "PRINT") {
					break
				}
				dtParts = append(dtParts, toks[i].String())
				// Handle parenthesized type args like VARCHAR(50)
				if toks[i].Kind == token.LParen {
					i++
					for i < len(toks) && toks[i].Kind != token.RParen {
						dtParts = append(dtParts, toks[i].String())
						i++
					}
					if i < len(toks) {
						dtParts = append(dtParts, toks[i].String())
					}
				}
				i++
			}
			sym.DataType = strings.Join(dtParts, "")
			st.Symbols = append(st.Symbols, sym)

			// Skip comma for next variable in the same DECLARE
			if i < len(toks) && toks[i].Kind == token.Comma {
				i++
				continue
			}
			break
		}
	}
}

// skipPastExpression advances past a value expression until comma or semicolon.
func skipPastExpression(toks []*ast.SQLToken, i int) int {
	depth := 0
	for i < len(toks) {
		if toks[i].Kind == token.LParen {
			depth++
		} else if toks[i].Kind == token.RParen {
			depth--
		}
		if depth == 0 && (toks[i].Kind == token.Comma || toks[i].Kind == token.Semicolon) {
			break
		}
		i++
	}
	return i
}

// extractCTEs handles: WITH name AS (...), name2 AS (...)
func extractCTEs(toks []*ast.SQLToken, st *SymbolTable) {
	for i := 0; i < len(toks); i++ {
		if !matchKeyword(toks[i], "WITH") {
			continue
		}
		// Check that the next non-WS token is an identifier, not a hint like (NOLOCK)
		j := skipWS(toks, i+1)
		if j >= len(toks) {
			continue
		}
		// If next token is '(' it's a table hint: WITH (NOLOCK)
		if toks[j].Kind == token.LParen {
			continue
		}

		// Parse CTE list: name AS (...) [, name AS (...)]
		i = j
		for i < len(toks) {
			i = skipWS(toks, i)
			if i >= len(toks) {
				break
			}
			// CTE name
			name := tokString(toks[i])
			if name == "" {
				break
			}
			sym := &Symbol{
				Name:   name,
				Kind:   SymbolCTE,
				Pos:    toks[i].From,
				EndPos: toks[i].To,
			}
			i++
			i = skipWS(toks, i)
			// Expect AS
			if i >= len(toks) || !matchKeyword(toks[i], "AS") {
				break
			}
			i++
			i = skipWS(toks, i)
			// Expect (
			if i >= len(toks) || toks[i].Kind != token.LParen {
				break
			}
			// Skip past parenthesized query
			depth := 1
			i++
			for i < len(toks) && depth > 0 {
				if toks[i].Kind == token.LParen {
					depth++
				} else if toks[i].Kind == token.RParen {
					depth--
				}
				i++
			}

			st.Symbols = append(st.Symbols, sym)

			// Check for comma (more CTEs)
			j := skipWS(toks, i)
			if j < len(toks) && toks[j].Kind == token.Comma {
				i = j + 1
				continue
			}
			break
		}
	}
}

// extractTempTables handles CREATE TABLE #name and SELECT ... INTO #name
func extractTempTables(toks []*ast.SQLToken, st *SymbolTable) {
	for i := 0; i < len(toks); i++ {
		// CREATE TABLE #tmp
		if matchKeyword(toks[i], "CREATE") {
			j := skipWS(toks, i+1)
			if j < len(toks) && matchKeyword(toks[j], "TABLE") {
				k := skipWS(toks, j+1)
				if k < len(toks) {
					name := tokString(toks[k])
					if strings.HasPrefix(name, "#") {
						st.Symbols = append(st.Symbols, &Symbol{
							Name:   name,
							Kind:   SymbolTempTable,
							Pos:    toks[k].From,
							EndPos: toks[k].To,
						})
					}
				}
			}
		}
		// SELECT ... INTO #tmp
		if matchKeyword(toks[i], "INTO") {
			j := skipWS(toks, i+1)
			if j < len(toks) {
				name := tokString(toks[j])
				if strings.HasPrefix(name, "#") {
					st.Symbols = append(st.Symbols, &Symbol{
						Name:   name,
						Kind:   SymbolTempTable,
						Pos:    toks[j].From,
						EndPos: toks[j].To,
					})
				}
			}
		}
	}
}
