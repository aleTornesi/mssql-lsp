package completer

import "github.com/aleTornesi/mssql-lsp/pkg/lsp"

type snippetDef struct {
	Label      string
	Detail     string
	InsertText string
}

var tsqlSnippets = []snippetDef{
	{
		Label:  "BEGIN...END",
		Detail: "Block",
		InsertText: `BEGIN
	$1
END$0`,
	},
	{
		Label:  "BEGIN TRY...END TRY BEGIN CATCH...END CATCH",
		Detail: "Try/Catch Block",
		InsertText: `BEGIN TRY
	$1
END TRY
BEGIN CATCH
	$2
END CATCH$0`,
	},
	{
		Label:  "IF...BEGIN...END ELSE BEGIN...END",
		Detail: "If/Else Block",
		InsertText: `IF $1
BEGIN
	$2
END
ELSE
BEGIN
	$3
END$0`,
	},
	{
		Label:  "WHILE...BEGIN...END",
		Detail: "While Loop",
		InsertText: `WHILE $1
BEGIN
	$2
END$0`,
	},
	{
		Label:  "DECLARE CURSOR",
		Detail: "Cursor Pattern",
		InsertText: `DECLARE $1 CURSOR FOR
$2

OPEN $1
FETCH NEXT FROM $1

WHILE @@FETCH_STATUS = 0
BEGIN
	$3
	FETCH NEXT FROM $1
END

CLOSE $1
DEALLOCATE $1$0`,
	},
}

func snippetCandidates() []lsp.CompletionItem {
	items := make([]lsp.CompletionItem, len(tsqlSnippets))
	for i, s := range tsqlSnippets {
		items[i] = lsp.CompletionItem{
			Label:           s.Label,
			Kind:            lsp.SnippetCompletion,
			Detail:          s.Detail,
			InsertText:      s.InsertText,
			InsertTextFormat: lsp.SnippetTextFormat,
		}
	}
	return items
}
