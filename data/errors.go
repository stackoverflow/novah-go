package data

import "fmt"

const (
	MISMATCHED_INDENTATION = "Mismatched indentation."

	MODULE_NAME = `Module names should be composed of identifiers started with a lower case character and separated by dots.
They also cannot contain special characters like '?' or '!'.`

	MODULE_DEFINITION = `Expected file to begin with a module declaration.
Example:

module some.package`

	IMPORT_REFER = "Expected exposing definitions to be a comma-separated list of upper or lower case identifiers."

	DECLARATION_REF_ALL = `To import or export all constructor of a type use a (..) syntax.

ex: import package (fun1, SomeType(..), fun2)`

	CTOR_NAME = "Expected constructor name (upper case identifier)."

	IMPORT_ALIAS = `Expected module import alias to be capitalized:
Example: import data.package as Mod`

	IMPORTED_DOT = "Expected identifier after imported variable reference."

	TYPE_VAR = "Expected type variable (lower case identifier)."

	TYPE_DEF = "Expected a type definition."

	TYPE_COLON = "Expected `:` before type definition."

	TYPEALIAS_DOT = "Expected type identifier after dot."

	TYPE_TEST_TYPE = "Expected type in type test."

	RECORD_LABEL = "A label of a record can only be a lower case identifier or a String."

	RECORD_COLON = "Expected `:` after record label."

	RECORD_EQUALS = "Expected `=` or `->` after record labels in set/update expression."

	INSTANCE_TYPE = "Instance types need to be enclosed in double brackets: {{ type }}."

	INSTANCE_VAR = "Instance variables need to be enclosed in double brackets: {{var}}."

	INSTANCE_ERROR = "Type and type alias declarations cannot be instances, only values."

	VARIABLE = "Expected variable name."

	OPERATOR = "Expected operator."

	LAMBDA_BACKSLASH = "Expected lambda definition to start with backslash: `\\`."

	LAMBDA_ARROW = "Expected `->` after lambda parameter definition."

	LAMBDA_VAR = `Expected identifier after start of lambda definition:
Example: \x -> x + 3`

	TOPLEVEL_IDENT = "Expected variable definition or variable type at the top level."

	PATTERN = `Expected a pattern expression.
|Patterns can be one of:
|
|Wildcard pattern: _
|Literal pattern: 3, 'a', "a string", false, etc
|Variable pattern: x, y, myVar, etc
|Constructor pattern: Some "ok", Result res, None, etc
|Record pattern: { x, y: 3 }
|List pattern: [], [x, y, _], [x :: xs]
|Named pattern: 10 as number
|Type test: :? Int as i`

	DO_WHILE = "Expected keyword `do` after while condition."

	EXP_SIMPLE = "Invalid expression for while condition."

	THEN = "Expected `then` after if condition."

	ELSE = "Expected `else` after then condition."

	LET_DECL = "Expected variable name after `let`."

	LET_EQUALS = "Expected `=` after let name declaration."

	LET_IN = "Expected `in` after let definition."

	FOR_IN = "Expected `in` after for pattern."

	FOR_DO = "Expected `do` after for definition."

	CASE_ARROW = "Expected `->` after case pattern."

	CASE_OF = "Expected `of` after a case expression."

	ALIAS_DOT = "Expected dot (.) after aliased variable."

	MALFORMED_EXPR = "Malformed expression."

	APPLIED_DO_LET = "Cannot apply let statement as a function."

	PUB_PLUS = "Visibility of value or typealias declaration can only be public (pub) not pub+."

	TYPEALIAS_NAME = "Expected name for typealias."

	TYPEALIAS_EQUALS = "Expected `=` after typealias declaration."

	DATA_NAME = "Expected new data type name to be a upper case identifier."

	DATA_EQUALS = "Expected equals `=` after data name declaration."

	INVALID_OPERATOR_DECL = "Operator declarations have to be defined between parentheses."
)

func UndefinedVarInCtor(name string, typeVars []string) string {
	if len(typeVars) == 1 {
		return fmt.Sprintf("The variable %s is undefined in constructor %s.", typeVars[0], name)
	}
	vars := JoinToStringFunc(typeVars, ", ", func(x string) string { return x })
	return fmt.Sprintf("The variables %s are undefined in constructor %s.", vars, name)
}

func ExpectedDefinition(name string) string {
	return fmt.Sprintf("Expected definition to follow its type declaration for %s.", name)
}

func ExpectedLetDefinition(name string) string {
	return fmt.Sprintf("Expected definition to follow its type declaration for %s in let clause.", name)
}

func EmptyImport(ctx string) string {
	return fmt.Sprintf("%s list cannot be empty.", ctx)
}

func WrongArityToCase(got int, expected int) string {
	return fmt.Sprintf("Case expression expected %d patterns but got %d.", got, expected)
}

func OpTooLong(op string) string {
	return fmt.Sprintf("Operator %s is too long. Operators cannot contain more than 3 characters.", op)
}

func LiteralExpected(name string) string {
	return fmt.Sprintf("Expected %s literal.", name)
}

func LParensExpected(ctx string) string {
	return fmt.Sprintf("Expected `(` after %s", ctx)
}

func RParensExpected(ctx string) string {
	return fmt.Sprintf("Expected `)` after %s", ctx)
}

func RSBracketExpected(ctx string) string {
	return fmt.Sprintf("Expected `]` after %s", ctx)
}

func RBracketExpected(ctx string) string {
	return fmt.Sprintf("Expected `}` after %s", ctx)
}

func PipeExpected(ctx string) string {
	return fmt.Sprintf("Expected `|` after %s.", ctx)
}

func CommaExpected(ctx string) string {
	return fmt.Sprintf("Expected `,` after %s.", ctx)
}

func EqualsExpected(ctx string) string {
	return fmt.Sprintf("Expected `=` after %s.", ctx)
}
