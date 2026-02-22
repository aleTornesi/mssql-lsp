package diagnostic

import (
	"strings"

	"github.com/atornesi/tsql-ls/ast"
	"github.com/atornesi/tsql-ls/token"
)

// CheckStructure walks the AST and reports structural errors.
func CheckStructure(root ast.TokenList, c *Collector) {
	walkTokenList(root, c)
}

func walkTokenList(tl ast.TokenList, c *Collector) {
	for _, node := range tl.GetTokens() {
		switch n := node.(type) {
		case *ast.BeginEnd:
			checkBeginEnd(n, c)
			walkTokenList(n, c)
		case *ast.TryCatch:
			checkTryCatch(n, c)
			walkTokenList(n, c)
		case *ast.Parenthesis:
			checkParenthesis(n, c)
			walkTokenList(n, c)
		case *ast.SwitchCase:
			checkSwitchCase(n, c)
			walkTokenList(n, c)
		case ast.TokenList:
			walkTokenList(n, c)
		}
	}
}

func checkBeginEnd(be *ast.BeginEnd, c *Collector) {
	toks := be.GetTokens()
	if len(toks) == 0 {
		return
	}
	last := toks[len(toks)-1]
	if !isKeywordNode(last, "END") {
		c.Add(Diagnostic{
			From:     be.Pos(),
			To:       endPos(be.Pos()),
			Severity: Error,
			Message:  "BEGIN without matching END",
			Code:     CodeUnclosedBegin,
		})
	}
}

func checkTryCatch(tc *ast.TryCatch, c *Collector) {
	toks := tc.GetTokens()
	if len(toks) == 0 {
		c.Add(Diagnostic{
			From:     tc.Pos(),
			To:       endPos(tc.Pos()),
			Severity: Error,
			Message:  "BEGIN TRY without matching END CATCH",
			Code:     CodeUnclosedTryCatch,
		})
		return
	}
	// Last non-whitespace token should be a MultiKeyword "END CATCH"
	found := false
	for i := len(toks) - 1; i >= 0; i-- {
		s := strings.TrimSpace(toks[i].String())
		if s == "" {
			continue
		}
		if strings.Contains(strings.ToUpper(s), "CATCH") {
			found = true
		}
		break
	}
	if !found {
		c.Add(Diagnostic{
			From:     tc.Pos(),
			To:       endPos(tc.Pos()),
			Severity: Error,
			Message:  "BEGIN TRY without matching END CATCH",
			Code:     CodeUnclosedTryCatch,
		})
	}
}

func checkParenthesis(p *ast.Parenthesis, c *Collector) {
	toks := p.GetTokens()
	if len(toks) == 0 {
		return
	}
	last := toks[len(toks)-1]
	if last.String() != ")" {
		c.Add(Diagnostic{
			From:     p.Pos(),
			To:       endPos(p.Pos()),
			Severity: Error,
			Message:  "unclosed parenthesis",
			Code:     CodeUnclosedParen,
		})
	}
}

func checkSwitchCase(sc *ast.SwitchCase, c *Collector) {
	toks := sc.GetTokens()
	if len(toks) == 0 {
		return
	}
	last := toks[len(toks)-1]
	if !isKeywordNode(last, "END") {
		c.Add(Diagnostic{
			From:     sc.Pos(),
			To:       endPos(sc.Pos()),
			Severity: Error,
			Message:  "unclosed CASE expression",
			Code:     CodeUnclosedCase,
		})
	}
}

func isKeywordNode(node ast.Node, kw string) bool {
	return strings.EqualFold(strings.TrimSpace(node.String()), kw)
}

// endPos returns a position one character past from, for point diagnostics.
func endPos(from token.Pos) token.Pos {
	return token.Pos{Line: from.Line, Col: from.Col + 1}
}
