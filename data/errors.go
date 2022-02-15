package data

import "fmt"

const MISMATCHED_INDENTATION = "Mismatched indentation."

const MODULE_NAME = `Module names should be composed of identifiers started with a lower case character and separated by dots.
They also cannot contain special characters like '?' or '!'.`

const MODULE_DEFINITION = `Expected file to begin with a module declaration.
Example:

module some.package`

const IMPORT_REFER = "Expected exposing definitions to be a comma-separated list of upper or lower case identifiers."

const DECLARATION_REF_ALL = `To import or export all constructor of a type use a (..) syntax.

ex: import package (fun1, SomeType(..), fun2)`

const CTOR_NAME = "Expected constructor name (upper case identifier)."

const IMPORT_ALIAS = `Expected module import alias to be capitalized:
Example: import data.package as Mod`

const IMPORTED_DOT = "Expected identifier after imported variable reference."

const TYPE_VAR = "Expected type variable (lower case identifier)."

const TYPE_DEF = "Expected a type definition."

const TYPE_COLON = "Expected `:` before type definition."

const TYPEALIAS_DOT = "Expected type identifier after dot."

const TYPE_TEST_TYPE = "Expected type in type test."

const RECORD_LABEL = "A label of a record can only be a lower case identifier or a String."

const RECORD_COLON = "Expected `:` after record label."

const INSTANCE_TYPE = "Instance types need to be enclosed in double brackets: {{ type }}."

const INSTANCE_VAR = "Instance variables need to be enclosed in double brackets: {{var}}."

const VARIABLE = "Expected variable name."

const OPERATOR = "Expected operator."

const LAMBDA_BACKSLASH = "Expected lambda definition to start with backslash: `\\`."

const LAMBDA_ARROW = "Expected `->` after lambda parameter definition."

const LAMBDA_VAR = `Expected identifier after start of lambda definition:
Example: \x -> x + 3`

const PATTERN = `Expected a pattern expression.
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

const DO_WHILE = "Expected keyword `do` after while condition."

const EXP_SIMPLE = "Invalid expression for while condition."

const THEN = "Expected `then` after if condition."

const ELSE = "Expected `else` after then condition."

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

func EmptyImport(ctx string) string {
	return fmt.Sprintf("%s list cannot be empty.", ctx)
}
