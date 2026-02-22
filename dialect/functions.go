package dialect

// BuiltinSignature describes a single overload of a built-in function.
type BuiltinSignature struct {
	Label      string
	Doc        string
	Parameters []BuiltinParam
}

// BuiltinParam describes a parameter of a built-in function.
type BuiltinParam struct {
	Label string
	Doc   string
}

// BuiltinFunction describes a built-in T-SQL function with its signatures.
type BuiltinFunction struct {
	Name       string
	Signatures []BuiltinSignature
}

// builtinFunctions is the registry of built-in function signatures.
var builtinFunctions = map[string]*BuiltinFunction{
	// Conversion functions
	"CAST": {
		Name: "CAST",
		Signatures: []BuiltinSignature{
			{Label: "CAST(expression AS data_type)", Doc: "Converts an expression to the specified data type.", Parameters: []BuiltinParam{
				{Label: "expression", Doc: "The value to convert."},
				{Label: "data_type", Doc: "The target data type."},
			}},
		},
	},
	"CONVERT": {
		Name: "CONVERT",
		Signatures: []BuiltinSignature{
			{Label: "CONVERT(data_type, expression [, style])", Doc: "Converts an expression to the specified data type.", Parameters: []BuiltinParam{
				{Label: "data_type", Doc: "The target data type."},
				{Label: "expression", Doc: "The value to convert."},
				{Label: "style", Doc: "Optional style for date/time formatting."},
			}},
		},
	},
	"TRY_CAST": {
		Name: "TRY_CAST",
		Signatures: []BuiltinSignature{
			{Label: "TRY_CAST(expression AS data_type)", Doc: "Like CAST but returns NULL on failure.", Parameters: []BuiltinParam{
				{Label: "expression"},
				{Label: "data_type"},
			}},
		},
	},
	"TRY_CONVERT": {
		Name: "TRY_CONVERT",
		Signatures: []BuiltinSignature{
			{Label: "TRY_CONVERT(data_type, expression [, style])", Doc: "Like CONVERT but returns NULL on failure.", Parameters: []BuiltinParam{
				{Label: "data_type"},
				{Label: "expression"},
				{Label: "style"},
			}},
		},
	},
	"PARSE": {
		Name: "PARSE",
		Signatures: []BuiltinSignature{
			{Label: "PARSE(string_value AS data_type [USING culture])", Doc: "Parses a string to the specified data type.", Parameters: []BuiltinParam{
				{Label: "string_value"},
				{Label: "data_type"},
				{Label: "culture"},
			}},
		},
	},
	"TRY_PARSE": {
		Name: "TRY_PARSE",
		Signatures: []BuiltinSignature{
			{Label: "TRY_PARSE(string_value AS data_type [USING culture])", Doc: "Like PARSE but returns NULL on failure.", Parameters: []BuiltinParam{
				{Label: "string_value"},
				{Label: "data_type"},
				{Label: "culture"},
			}},
		},
	},

	// String functions
	"SUBSTRING": {
		Name: "SUBSTRING",
		Signatures: []BuiltinSignature{
			{Label: "SUBSTRING(expression, start, length)", Doc: "Returns part of a string.", Parameters: []BuiltinParam{
				{Label: "expression", Doc: "The source string."},
				{Label: "start", Doc: "Starting position (1-based)."},
				{Label: "length", Doc: "Number of characters."},
			}},
		},
	},
	"CHARINDEX": {
		Name: "CHARINDEX",
		Signatures: []BuiltinSignature{
			{Label: "CHARINDEX(substring, string [, start])", Doc: "Returns the position of a substring.", Parameters: []BuiltinParam{
				{Label: "substring"},
				{Label: "string"},
				{Label: "start"},
			}},
		},
	},
	"REPLACE": {
		Name: "REPLACE",
		Signatures: []BuiltinSignature{
			{Label: "REPLACE(string, old, new)", Doc: "Replaces all occurrences of a substring.", Parameters: []BuiltinParam{
				{Label: "string"},
				{Label: "old"},
				{Label: "new"},
			}},
		},
	},
	"STUFF": {
		Name: "STUFF",
		Signatures: []BuiltinSignature{
			{Label: "STUFF(string, start, length, replacement)", Doc: "Deletes and inserts characters.", Parameters: []BuiltinParam{
				{Label: "string"},
				{Label: "start"},
				{Label: "length"},
				{Label: "replacement"},
			}},
		},
	},
	"LEFT": {
		Name: "LEFT",
		Signatures: []BuiltinSignature{
			{Label: "LEFT(string, count)", Doc: "Returns leftmost characters.", Parameters: []BuiltinParam{
				{Label: "string"},
				{Label: "count"},
			}},
		},
	},
	"RIGHT": {
		Name: "RIGHT",
		Signatures: []BuiltinSignature{
			{Label: "RIGHT(string, count)", Doc: "Returns rightmost characters.", Parameters: []BuiltinParam{
				{Label: "string"},
				{Label: "count"},
			}},
		},
	},
	"LEN": {
		Name: "LEN",
		Signatures: []BuiltinSignature{
			{Label: "LEN(string)", Doc: "Returns the length of a string.", Parameters: []BuiltinParam{
				{Label: "string"},
			}},
		},
	},
	"LTRIM": {
		Name: "LTRIM",
		Signatures: []BuiltinSignature{
			{Label: "LTRIM(string)", Doc: "Removes leading blanks.", Parameters: []BuiltinParam{
				{Label: "string"},
			}},
		},
	},
	"RTRIM": {
		Name: "RTRIM",
		Signatures: []BuiltinSignature{
			{Label: "RTRIM(string)", Doc: "Removes trailing blanks.", Parameters: []BuiltinParam{
				{Label: "string"},
			}},
		},
	},
	"TRIM": {
		Name: "TRIM",
		Signatures: []BuiltinSignature{
			{Label: "TRIM(string)", Doc: "Removes leading and trailing blanks.", Parameters: []BuiltinParam{
				{Label: "string"},
			}},
		},
	},
	"UPPER": {
		Name: "UPPER",
		Signatures: []BuiltinSignature{
			{Label: "UPPER(string)", Doc: "Converts to uppercase.", Parameters: []BuiltinParam{
				{Label: "string"},
			}},
		},
	},
	"LOWER": {
		Name: "LOWER",
		Signatures: []BuiltinSignature{
			{Label: "LOWER(string)", Doc: "Converts to lowercase.", Parameters: []BuiltinParam{
				{Label: "string"},
			}},
		},
	},
	"CONCAT": {
		Name: "CONCAT",
		Signatures: []BuiltinSignature{
			{Label: "CONCAT(value1, value2 [, ...])", Doc: "Concatenates values.", Parameters: []BuiltinParam{
				{Label: "value1"},
				{Label: "value2"},
			}},
		},
	},
	"CONCAT_WS": {
		Name: "CONCAT_WS",
		Signatures: []BuiltinSignature{
			{Label: "CONCAT_WS(separator, value1, value2 [, ...])", Doc: "Concatenates with separator.", Parameters: []BuiltinParam{
				{Label: "separator"},
				{Label: "value1"},
				{Label: "value2"},
			}},
		},
	},
	"FORMAT": {
		Name: "FORMAT",
		Signatures: []BuiltinSignature{
			{Label: "FORMAT(value, format [, culture])", Doc: "Formats a value with a .NET format string.", Parameters: []BuiltinParam{
				{Label: "value"},
				{Label: "format"},
				{Label: "culture"},
			}},
		},
	},
	"REPLICATE": {
		Name: "REPLICATE",
		Signatures: []BuiltinSignature{
			{Label: "REPLICATE(string, count)", Doc: "Repeats a string.", Parameters: []BuiltinParam{
				{Label: "string"},
				{Label: "count"},
			}},
		},
	},
	"REVERSE": {
		Name: "REVERSE",
		Signatures: []BuiltinSignature{
			{Label: "REVERSE(string)", Doc: "Reverses a string.", Parameters: []BuiltinParam{
				{Label: "string"},
			}},
		},
	},
	"STRING_AGG": {
		Name: "STRING_AGG",
		Signatures: []BuiltinSignature{
			{Label: "STRING_AGG(expression, separator)", Doc: "Concatenates values with a separator.", Parameters: []BuiltinParam{
				{Label: "expression"},
				{Label: "separator"},
			}},
		},
	},
	"STRING_SPLIT": {
		Name: "STRING_SPLIT",
		Signatures: []BuiltinSignature{
			{Label: "STRING_SPLIT(string, separator)", Doc: "Splits a string into rows.", Parameters: []BuiltinParam{
				{Label: "string"},
				{Label: "separator"},
			}},
		},
	},
	"TRANSLATE": {
		Name: "TRANSLATE",
		Signatures: []BuiltinSignature{
			{Label: "TRANSLATE(string, from_chars, to_chars)", Doc: "Replaces characters.", Parameters: []BuiltinParam{
				{Label: "string"},
				{Label: "from_chars"},
				{Label: "to_chars"},
			}},
		},
	},
	"QUOTENAME": {
		Name: "QUOTENAME",
		Signatures: []BuiltinSignature{
			{Label: "QUOTENAME(string [, quote_char])", Doc: "Returns a string with delimiters.", Parameters: []BuiltinParam{
				{Label: "string"},
				{Label: "quote_char"},
			}},
		},
	},

	// Date/time functions
	"DATEADD": {
		Name: "DATEADD",
		Signatures: []BuiltinSignature{
			{Label: "DATEADD(datepart, number, date)", Doc: "Adds an interval to a date.", Parameters: []BuiltinParam{
				{Label: "datepart", Doc: "The part of date to add to (year, month, day, etc.)."},
				{Label: "number", Doc: "The value to add."},
				{Label: "date", Doc: "The date to modify."},
			}},
		},
	},
	"DATEDIFF": {
		Name: "DATEDIFF",
		Signatures: []BuiltinSignature{
			{Label: "DATEDIFF(datepart, startdate, enddate)", Doc: "Returns the difference between two dates.", Parameters: []BuiltinParam{
				{Label: "datepart"},
				{Label: "startdate"},
				{Label: "enddate"},
			}},
		},
	},
	"DATEDIFF_BIG": {
		Name: "DATEDIFF_BIG",
		Signatures: []BuiltinSignature{
			{Label: "DATEDIFF_BIG(datepart, startdate, enddate)", Doc: "Like DATEDIFF but returns bigint.", Parameters: []BuiltinParam{
				{Label: "datepart"},
				{Label: "startdate"},
				{Label: "enddate"},
			}},
		},
	},
	"DATENAME": {
		Name: "DATENAME",
		Signatures: []BuiltinSignature{
			{Label: "DATENAME(datepart, date)", Doc: "Returns a date part as a string.", Parameters: []BuiltinParam{
				{Label: "datepart"},
				{Label: "date"},
			}},
		},
	},
	"DATEPART": {
		Name: "DATEPART",
		Signatures: []BuiltinSignature{
			{Label: "DATEPART(datepart, date)", Doc: "Returns a date part as an integer.", Parameters: []BuiltinParam{
				{Label: "datepart"},
				{Label: "date"},
			}},
		},
	},
	"DATETRUNC": {
		Name: "DATETRUNC",
		Signatures: []BuiltinSignature{
			{Label: "DATETRUNC(datepart, date)", Doc: "Truncates a date to the specified precision.", Parameters: []BuiltinParam{
				{Label: "datepart"},
				{Label: "date"},
			}},
		},
	},
	"DATEFROMPARTS": {
		Name: "DATEFROMPARTS",
		Signatures: []BuiltinSignature{
			{Label: "DATEFROMPARTS(year, month, day)", Doc: "Returns a date from parts.", Parameters: []BuiltinParam{
				{Label: "year"},
				{Label: "month"},
				{Label: "day"},
			}},
		},
	},
	"EOMONTH": {
		Name: "EOMONTH",
		Signatures: []BuiltinSignature{
			{Label: "EOMONTH(start_date [, month_to_add])", Doc: "Returns last day of the month.", Parameters: []BuiltinParam{
				{Label: "start_date"},
				{Label: "month_to_add"},
			}},
		},
	},
	"SWITCHOFFSET": {
		Name: "SWITCHOFFSET",
		Signatures: []BuiltinSignature{
			{Label: "SWITCHOFFSET(datetimeoffset, time_zone)", Doc: "Changes the time zone offset.", Parameters: []BuiltinParam{
				{Label: "datetimeoffset"},
				{Label: "time_zone"},
			}},
		},
	},
	"TODATETIMEOFFSET": {
		Name: "TODATETIMEOFFSET",
		Signatures: []BuiltinSignature{
			{Label: "TODATETIMEOFFSET(expression, time_zone)", Doc: "Adds a time zone offset.", Parameters: []BuiltinParam{
				{Label: "expression"},
				{Label: "time_zone"},
			}},
		},
	},
	"ISDATE": {
		Name: "ISDATE",
		Signatures: []BuiltinSignature{
			{Label: "ISDATE(expression)", Doc: "Returns 1 if valid date/time.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},
	"GETDATE": {
		Name: "GETDATE",
		Signatures: []BuiltinSignature{
			{Label: "GETDATE()", Doc: "Returns current date and time."},
		},
	},
	"GETUTCDATE": {
		Name: "GETUTCDATE",
		Signatures: []BuiltinSignature{
			{Label: "GETUTCDATE()", Doc: "Returns current UTC date and time."},
		},
	},
	"SYSDATETIME": {
		Name: "SYSDATETIME",
		Signatures: []BuiltinSignature{
			{Label: "SYSDATETIME()", Doc: "Returns current date and time with higher precision."},
		},
	},
	"SYSUTCDATETIME": {
		Name: "SYSUTCDATETIME",
		Signatures: []BuiltinSignature{
			{Label: "SYSUTCDATETIME()", Doc: "Returns current UTC date and time with higher precision."},
		},
	},

	// Logical / conditional functions
	"COALESCE": {
		Name: "COALESCE",
		Signatures: []BuiltinSignature{
			{Label: "COALESCE(expression1, expression2 [, ...])", Doc: "Returns first non-NULL expression.", Parameters: []BuiltinParam{
				{Label: "expression1"},
				{Label: "expression2"},
			}},
		},
	},
	"NULLIF": {
		Name: "NULLIF",
		Signatures: []BuiltinSignature{
			{Label: "NULLIF(expression1, expression2)", Doc: "Returns NULL if the two expressions are equal.", Parameters: []BuiltinParam{
				{Label: "expression1"},
				{Label: "expression2"},
			}},
		},
	},
	"ISNULL": {
		Name: "ISNULL",
		Signatures: []BuiltinSignature{
			{Label: "ISNULL(check_expression, replacement_value)", Doc: "Replaces NULL with a specified value.", Parameters: []BuiltinParam{
				{Label: "check_expression"},
				{Label: "replacement_value"},
			}},
		},
	},
	"IIF": {
		Name: "IIF",
		Signatures: []BuiltinSignature{
			{Label: "IIF(boolean_expression, true_value, false_value)", Doc: "Returns one of two values based on a condition.", Parameters: []BuiltinParam{
				{Label: "boolean_expression"},
				{Label: "true_value"},
				{Label: "false_value"},
			}},
		},
	},
	"CHOOSE": {
		Name: "CHOOSE",
		Signatures: []BuiltinSignature{
			{Label: "CHOOSE(index, val1, val2 [, ...])", Doc: "Returns item at specified index.", Parameters: []BuiltinParam{
				{Label: "index"},
				{Label: "val1"},
				{Label: "val2"},
			}},
		},
	},

	// Math functions
	"ABS": {
		Name: "ABS",
		Signatures: []BuiltinSignature{
			{Label: "ABS(numeric_expression)", Doc: "Returns the absolute value.", Parameters: []BuiltinParam{
				{Label: "numeric_expression"},
			}},
		},
	},
	"CEILING": {
		Name: "CEILING",
		Signatures: []BuiltinSignature{
			{Label: "CEILING(numeric_expression)", Doc: "Returns the smallest integer >= the value.", Parameters: []BuiltinParam{
				{Label: "numeric_expression"},
			}},
		},
	},
	"FLOOR": {
		Name: "FLOOR",
		Signatures: []BuiltinSignature{
			{Label: "FLOOR(numeric_expression)", Doc: "Returns the largest integer <= the value.", Parameters: []BuiltinParam{
				{Label: "numeric_expression"},
			}},
		},
	},
	"ROUND": {
		Name: "ROUND",
		Signatures: []BuiltinSignature{
			{Label: "ROUND(numeric_expression, length [, function])", Doc: "Rounds a numeric value.", Parameters: []BuiltinParam{
				{Label: "numeric_expression"},
				{Label: "length"},
				{Label: "function"},
			}},
		},
	},
	"POWER": {
		Name: "POWER",
		Signatures: []BuiltinSignature{
			{Label: "POWER(float_expression, y)", Doc: "Returns the value raised to a power.", Parameters: []BuiltinParam{
				{Label: "float_expression"},
				{Label: "y"},
			}},
		},
	},
	"SQRT": {
		Name: "SQRT",
		Signatures: []BuiltinSignature{
			{Label: "SQRT(float_expression)", Doc: "Returns the square root.", Parameters: []BuiltinParam{
				{Label: "float_expression"},
			}},
		},
	},
	"RAND": {
		Name: "RAND",
		Signatures: []BuiltinSignature{
			{Label: "RAND([seed])", Doc: "Returns a pseudo-random float value.", Parameters: []BuiltinParam{
				{Label: "seed"},
			}},
		},
	},
	"LOG": {
		Name: "LOG",
		Signatures: []BuiltinSignature{
			{Label: "LOG(float_expression [, base])", Doc: "Returns the natural logarithm.", Parameters: []BuiltinParam{
				{Label: "float_expression"},
				{Label: "base"},
			}},
		},
	},

	// Aggregate functions
	"COUNT": {
		Name: "COUNT",
		Signatures: []BuiltinSignature{
			{Label: "COUNT(expression)", Doc: "Returns the number of items.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},
	"COUNT_BIG": {
		Name: "COUNT_BIG",
		Signatures: []BuiltinSignature{
			{Label: "COUNT_BIG(expression)", Doc: "Like COUNT but returns bigint.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},
	"SUM": {
		Name: "SUM",
		Signatures: []BuiltinSignature{
			{Label: "SUM(expression)", Doc: "Returns the sum.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},
	"AVG": {
		Name: "AVG",
		Signatures: []BuiltinSignature{
			{Label: "AVG(expression)", Doc: "Returns the average.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},
	"MIN": {
		Name: "MIN",
		Signatures: []BuiltinSignature{
			{Label: "MIN(expression)", Doc: "Returns the minimum value.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},
	"MAX": {
		Name: "MAX",
		Signatures: []BuiltinSignature{
			{Label: "MAX(expression)", Doc: "Returns the maximum value.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},

	// Window functions
	"ROW_NUMBER": {
		Name: "ROW_NUMBER",
		Signatures: []BuiltinSignature{
			{Label: "ROW_NUMBER()", Doc: "Returns a sequential row number."},
		},
	},
	"RANK": {
		Name: "RANK",
		Signatures: []BuiltinSignature{
			{Label: "RANK()", Doc: "Returns the rank with gaps."},
		},
	},
	"DENSE_RANK": {
		Name: "DENSE_RANK",
		Signatures: []BuiltinSignature{
			{Label: "DENSE_RANK()", Doc: "Returns the rank without gaps."},
		},
	},
	"NTILE": {
		Name: "NTILE",
		Signatures: []BuiltinSignature{
			{Label: "NTILE(integer_expression)", Doc: "Distributes rows into groups.", Parameters: []BuiltinParam{
				{Label: "integer_expression"},
			}},
		},
	},
	"LAG": {
		Name: "LAG",
		Signatures: []BuiltinSignature{
			{Label: "LAG(expression [, offset [, default]])", Doc: "Accesses a previous row.", Parameters: []BuiltinParam{
				{Label: "expression"},
				{Label: "offset"},
				{Label: "default"},
			}},
		},
	},
	"LEAD": {
		Name: "LEAD",
		Signatures: []BuiltinSignature{
			{Label: "LEAD(expression [, offset [, default]])", Doc: "Accesses a subsequent row.", Parameters: []BuiltinParam{
				{Label: "expression"},
				{Label: "offset"},
				{Label: "default"},
			}},
		},
	},
	"FIRST_VALUE": {
		Name: "FIRST_VALUE",
		Signatures: []BuiltinSignature{
			{Label: "FIRST_VALUE(expression)", Doc: "Returns first value in an ordered set.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},
	"LAST_VALUE": {
		Name: "LAST_VALUE",
		Signatures: []BuiltinSignature{
			{Label: "LAST_VALUE(expression)", Doc: "Returns last value in an ordered set.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},

	// System functions
	"DATALENGTH": {
		Name: "DATALENGTH",
		Signatures: []BuiltinSignature{
			{Label: "DATALENGTH(expression)", Doc: "Returns bytes used to represent expression.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},
	"DB_ID": {
		Name: "DB_ID",
		Signatures: []BuiltinSignature{
			{Label: "DB_ID([database_name])", Doc: "Returns the database ID.", Parameters: []BuiltinParam{
				{Label: "database_name"},
			}},
		},
	},
	"DB_NAME": {
		Name: "DB_NAME",
		Signatures: []BuiltinSignature{
			{Label: "DB_NAME([database_id])", Doc: "Returns the database name.", Parameters: []BuiltinParam{
				{Label: "database_id"},
			}},
		},
	},
	"OBJECT_ID": {
		Name: "OBJECT_ID",
		Signatures: []BuiltinSignature{
			{Label: "OBJECT_ID(object_name [, object_type])", Doc: "Returns the object ID.", Parameters: []BuiltinParam{
				{Label: "object_name"},
				{Label: "object_type"},
			}},
		},
	},
	"OBJECT_NAME": {
		Name: "OBJECT_NAME",
		Signatures: []BuiltinSignature{
			{Label: "OBJECT_NAME(object_id [, database_id])", Doc: "Returns the object name.", Parameters: []BuiltinParam{
				{Label: "object_id"},
				{Label: "database_id"},
			}},
		},
	},
	"SCHEMA_NAME": {
		Name: "SCHEMA_NAME",
		Signatures: []BuiltinSignature{
			{Label: "SCHEMA_NAME([schema_id])", Doc: "Returns the schema name.", Parameters: []BuiltinParam{
				{Label: "schema_id"},
			}},
		},
	},
	"SCOPE_IDENTITY": {
		Name: "SCOPE_IDENTITY",
		Signatures: []BuiltinSignature{
			{Label: "SCOPE_IDENTITY()", Doc: "Returns the last identity value inserted."},
		},
	},
	"NEWID": {
		Name: "NEWID",
		Signatures: []BuiltinSignature{
			{Label: "NEWID()", Doc: "Returns a new uniqueidentifier value."},
		},
	},
	"ISNUMERIC": {
		Name: "ISNUMERIC",
		Signatures: []BuiltinSignature{
			{Label: "ISNUMERIC(expression)", Doc: "Returns 1 if the expression is a valid numeric type.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},

	// Error functions
	"ERROR_MESSAGE": {
		Name: "ERROR_MESSAGE",
		Signatures: []BuiltinSignature{
			{Label: "ERROR_MESSAGE()", Doc: "Returns the message text of the error."},
		},
	},
	"ERROR_NUMBER": {
		Name: "ERROR_NUMBER",
		Signatures: []BuiltinSignature{
			{Label: "ERROR_NUMBER()", Doc: "Returns the error number."},
		},
	},
	"ERROR_LINE": {
		Name: "ERROR_LINE",
		Signatures: []BuiltinSignature{
			{Label: "ERROR_LINE()", Doc: "Returns the line number where error occurred."},
		},
	},
	"ERROR_SEVERITY": {
		Name: "ERROR_SEVERITY",
		Signatures: []BuiltinSignature{
			{Label: "ERROR_SEVERITY()", Doc: "Returns the severity of the error."},
		},
	},
	"ERROR_STATE": {
		Name: "ERROR_STATE",
		Signatures: []BuiltinSignature{
			{Label: "ERROR_STATE()", Doc: "Returns the state number of the error."},
		},
	},
	"ERROR_PROCEDURE": {
		Name: "ERROR_PROCEDURE",
		Signatures: []BuiltinSignature{
			{Label: "ERROR_PROCEDURE()", Doc: "Returns the name of the stored procedure where error occurred."},
		},
	},

	// JSON functions
	"JSON_VALUE": {
		Name: "JSON_VALUE",
		Signatures: []BuiltinSignature{
			{Label: "JSON_VALUE(expression, path)", Doc: "Extracts a scalar value from a JSON string.", Parameters: []BuiltinParam{
				{Label: "expression"},
				{Label: "path"},
			}},
		},
	},
	"JSON_QUERY": {
		Name: "JSON_QUERY",
		Signatures: []BuiltinSignature{
			{Label: "JSON_QUERY(expression [, path])", Doc: "Extracts an object or array from JSON.", Parameters: []BuiltinParam{
				{Label: "expression"},
				{Label: "path"},
			}},
		},
	},
	"JSON_MODIFY": {
		Name: "JSON_MODIFY",
		Signatures: []BuiltinSignature{
			{Label: "JSON_MODIFY(expression, path, newValue)", Doc: "Updates a value in a JSON string.", Parameters: []BuiltinParam{
				{Label: "expression"},
				{Label: "path"},
				{Label: "newValue"},
			}},
		},
	},
	"ISJSON": {
		Name: "ISJSON",
		Signatures: []BuiltinSignature{
			{Label: "ISJSON(expression)", Doc: "Tests whether a string is valid JSON.", Parameters: []BuiltinParam{
				{Label: "expression"},
			}},
		},
	},
	"OPENJSON": {
		Name: "OPENJSON",
		Signatures: []BuiltinSignature{
			{Label: "OPENJSON(jsonExpression [, path])", Doc: "Parses JSON text and returns objects and properties as rows.", Parameters: []BuiltinParam{
				{Label: "jsonExpression"},
				{Label: "path"},
			}},
		},
	},

	// Cryptographic
	"HASHBYTES": {
		Name: "HASHBYTES",
		Signatures: []BuiltinSignature{
			{Label: "HASHBYTES(algorithm, input)", Doc: "Returns the hash of the input.", Parameters: []BuiltinParam{
				{Label: "algorithm", Doc: "'MD2', 'MD4', 'MD5', 'SHA', 'SHA1', 'SHA2_256', 'SHA2_512'"},
				{Label: "input"},
			}},
		},
	},

	// Other
	"FORMATMESSAGE": {
		Name: "FORMATMESSAGE",
		Signatures: []BuiltinSignature{
			{Label: "FORMATMESSAGE(msg_number | msg_string, param_value [, ...])", Doc: "Constructs a message from an existing message.", Parameters: []BuiltinParam{
				{Label: "msg_number | msg_string"},
				{Label: "param_value"},
			}},
		},
	},
	"XACT_STATE": {
		Name: "XACT_STATE",
		Signatures: []BuiltinSignature{
			{Label: "XACT_STATE()", Doc: "Reports the transaction state of the current session."},
		},
	},
}

// LookupBuiltinFunction returns the signature info for a built-in function.
func LookupBuiltinFunction(name string) *BuiltinFunction {
	return builtinFunctions[name]
}
