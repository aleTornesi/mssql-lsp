package diagnostic

import "github.com/atornesi/tsql-ls/token"

// Severity matches LSP DiagnosticSeverity.
type Severity int

const (
	Error   Severity = 1
	Warning Severity = 2
	Info    Severity = 3
	Hint    Severity = 4
)

// Diagnostic codes.
const (
	CodeUnclosedComment   = "TSQL001"
	CodeUnclosedString    = "TSQL002"
	CodeUnclosedBracket   = "TSQL003"
	CodeUnclosedBegin     = "TSQL004"
	CodeUnclosedTryCatch  = "TSQL005"
	CodeUnclosedParen     = "TSQL006"
	CodeUnclosedCase      = "TSQL007"
	CodeIllegalChar       = "TSQL008"
	CodeDuplicateVariable = "TSQL010"
	CodeUnreferencedCTE   = "TSQL011"
)

// Diagnostic represents a single diagnostic finding.
type Diagnostic struct {
	From     token.Pos
	To       token.Pos
	Severity Severity
	Message  string
	Code     string
}

// Collector accumulates diagnostics.
type Collector struct {
	Diagnostics []Diagnostic
}

// Add appends a diagnostic.
func (c *Collector) Add(d Diagnostic) {
	c.Diagnostics = append(c.Diagnostics, d)
}

// Merge appends all diagnostics from another collector.
func (c *Collector) Merge(other *Collector) {
	c.Diagnostics = append(c.Diagnostics, other.Diagnostics...)
}

// OffsetLines shifts all diagnostic positions by startLine.
func (c *Collector) OffsetLines(startLine int) {
	for i := range c.Diagnostics {
		c.Diagnostics[i].From.Line += startLine
		c.Diagnostics[i].To.Line += startLine
	}
}
