package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/atornesi/tsql-ls/ast"
	"github.com/atornesi/tsql-ls/ast/astutil"
	"github.com/atornesi/tsql-ls/dialect"
	"github.com/atornesi/tsql-ls/internal/database"
	"github.com/atornesi/tsql-ls/internal/lsp"
	"github.com/atornesi/tsql-ls/parser"
	"github.com/atornesi/tsql-ls/parser/parseutil"
	"github.com/atornesi/tsql-ls/token"
)

func (s *Server) handleTextDocumentSignatureHelp(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.SignatureHelpParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	f, ok := s.files[params.TextDocument.URI]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	res, err := SignatureHelp(f.Text, params, s.worker.Cache())
	if err != nil {
		return nil, err
	}
	return res, nil
}

func SignatureHelp(text string, params lsp.SignatureHelpParams, dbCache *database.DBCache) (*lsp.SignatureHelp, error) {
	batchText, adjustedLine := parser.BatchAtLine(text, params.Position.Line)
	parsed, err := parser.Parse(batchText)
	if err != nil {
		return nil, err
	}

	pos := token.Pos{
		Line: adjustedLine,
		Col:  params.Position.Character,
	}
	nodeWalker := parseutil.NewNodeWalker(parsed, pos)
	types := getSignatureHelpTypes(nodeWalker)

	switch {
	case signatureHelpIs(types, SignatureHelpTypeInsertValue) && dbCache != nil:
		insert, err := parseutil.ExtractInsert(parsed, pos)
		if err != nil {
			return nil, err
		}
		if !insert.Enable() {
			return nil, err
		}

		table := insert.GetTable()
		cols := insert.GetColumns()
		paramIdx := insert.GetValues().GetIndex(pos)
		tableName := table.Name

		params := []lsp.ParameterInformation{}
		for _, col := range cols.GetIdentifiers() {
			colName := col.String()
			colDoc := ""
			colDesc, ok := dbCache.Column(tableName, colName)
			if ok {
				colDoc = colDesc.OnelineDesc()
			}
			p := lsp.ParameterInformation{
				Label:         colName,
				Documentation: colDoc,
			}
			params = append(params, p)
		}

		signatureLabel := fmt.Sprintf("%s (%s)", tableName, cols.String())
		sh := &lsp.SignatureHelp{
			Signatures: []lsp.SignatureInformation{
				{
					Label:         signatureLabel,
					Documentation: fmt.Sprintf("%s table columns", tableName),
					Parameters:    params,
				},
			},
			ActiveSignature: 0.0,
			ActiveParameter: float64(paramIdx),
		}
		return sh, nil
	default:
		// Try builtin function signature help
		return builtinSignatureHelp(nodeWalker, pos)
	}
}

func builtinSignatureHelp(nw *parseutil.NodeWalker, pos token.Pos) (*lsp.SignatureHelp, error) {
	// Find enclosing FunctionLiteral
	m := astutil.NodeMatcher{NodeTypes: []ast.NodeType{ast.TypeFunctionLiteral}}
	fnNode := nw.CurNodeBottomMatched(m)
	if fnNode == nil {
		return nil, nil
	}

	fl, ok := fnNode.(*ast.FunctionLiteral)
	if !ok {
		return nil, nil
	}

	// Get function name from first token
	tokens := fl.GetTokens()
	if len(tokens) == 0 {
		return nil, nil
	}

	funcName := ""
	if item, ok := tokens[0].(ast.Token); ok {
		funcName = strings.ToUpper(item.GetToken().NoQuoteString())
	} else {
		funcName = strings.ToUpper(tokens[0].String())
	}

	builtin := dialect.LookupBuiltinFunction(funcName)
	if builtin == nil {
		return nil, nil
	}

	// Count commas before cursor position to determine active parameter
	activeParam := countCommasBeforePos(fl, pos)

	sigs := make([]lsp.SignatureInformation, len(builtin.Signatures))
	for i, sig := range builtin.Signatures {
		params := make([]lsp.ParameterInformation, len(sig.Parameters))
		for j, p := range sig.Parameters {
			params[j] = lsp.ParameterInformation{
				Label:         p.Label,
				Documentation: p.Doc,
			}
		}
		sigs[i] = lsp.SignatureInformation{
			Label:         sig.Label,
			Documentation: sig.Doc,
			Parameters:    params,
		}
	}

	return &lsp.SignatureHelp{
		Signatures:      sigs,
		ActiveSignature: 0,
		ActiveParameter: float64(activeParam),
	}, nil
}

func countCommasBeforePos(fl *ast.FunctionLiteral, pos token.Pos) int {
	count := 0
	for _, node := range fl.GetTokens() {
		if token.ComparePos(node.Pos(), pos) >= 0 {
			break
		}
		if item, ok := node.(ast.Token); ok {
			if item.GetToken().MatchKind(token.Comma) {
				count++
			}
		}
		// Check inside parenthesis
		if paren, ok := node.(*ast.Parenthesis); ok {
			for _, inner := range paren.GetTokens() {
				if token.ComparePos(inner.Pos(), pos) >= 0 {
					break
				}
				if item, ok := inner.(ast.Token); ok {
					if item.GetToken().MatchKind(token.Comma) {
						count++
					}
				}
			}
		}
	}
	return count
}

type signatureHelpType int

const (
	_ signatureHelpType = iota
	SignatureHelpTypeInsertValue
	SignatureHelpTypeUnknown = 99
)

func (sht signatureHelpType) String() string {
	switch sht {
	case SignatureHelpTypeInsertValue:
		return "InsertValue"
	default:
		return ""
	}
}

func getSignatureHelpTypes(nw *parseutil.NodeWalker) []signatureHelpType {
	syntaxPos := parseutil.CheckSyntaxPosition(nw)
	types := []signatureHelpType{}
	switch {
	case syntaxPos == parseutil.InsertValue:
		types = []signatureHelpType{
			SignatureHelpTypeInsertValue,
		}
	default:
		// pass
	}
	return types
}

func signatureHelpIs(types []signatureHelpType, expect signatureHelpType) bool {
	for _, t := range types {
		if t == expect {
			return true
		}
	}
	return false
}
