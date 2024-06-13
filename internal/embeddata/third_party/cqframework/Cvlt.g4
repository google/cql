grammar Cvlt;

/*
This is a literal grammar for CQL that supports all literals from CQL, including type specifiers
Based on FHIRPath and CQL grammars for CQL v1.5.2
 */

/*
 * Type Specifiers
 */

typeSpecifier
    : namedTypeSpecifier
    | listTypeSpecifier
    | intervalTypeSpecifier
    | tupleTypeSpecifier
    | choiceTypeSpecifier
    ;

namedTypeSpecifier
    : (identifier '.')* identifier
    ;

modelIdentifier
    : identifier
    ;

listTypeSpecifier
    : 'List' '<' typeSpecifier '>'
    ;

intervalTypeSpecifier
    : 'Interval' '<' typeSpecifier '>'
    ;

tupleTypeSpecifier
    : 'Tuple' '{' tupleElementDefinition (',' tupleElementDefinition)* '}'
    ;

tupleElementDefinition
    : identifier typeSpecifier
    ;

choiceTypeSpecifier
    : 'Choice' '<' typeSpecifier (',' typeSpecifier)* '>'
    ;

term
    : literal               #literalTerm
    | intervalSelector      #intervalSelectorTerm
    | tupleSelector         #tupleSelectorTerm
    | instanceSelector      #instanceSelectorTerm
    | listSelector          #listSelectorTerm
    | codeSelector          #codeSelectorTerm
    | conceptSelector       #conceptSelectorTerm
    ;

ratio
    : quantity ':' quantity
    ;

literal
    : ('true' | 'false')                                    #booleanLiteral
    | 'null'                                                #nullLiteral
    | STRING                                                #stringLiteral
    | NUMBER                                                #numberLiteral
    | LONGNUMBER                                            #longNumberLiteral
    | DATETIME                                              #dateTimeLiteral
    | DATE                                                  #dateLiteral
    | TIME                                                  #timeLiteral
    | quantity                                              #quantityLiteral
    | ratio                                                 #ratioLiteral
    ;

intervalSelector
    : 'Interval' ('['|'(') literal ',' literal (']'|')')
    ;

tupleSelector
    : 'Tuple'? '{' (':' | (tupleElementSelector (',' tupleElementSelector)*)) '}'
    ;

tupleElementSelector
    : identifier ':' term
    ;

instanceSelector
    : identifier '{' (':' | (instanceElementSelector (',' instanceElementSelector)*)) '}'
    ;

instanceElementSelector
    : identifier ':' term
    ;

listSelector
    : ('List' ('<' typeSpecifier '>')?)? '{' (term (',' term)*)? '}'
    ;

displayClause
    : 'display' STRING
    ;

codeSelector
    : 'Code' STRING 'from' identifier displayClause?
    ;

conceptSelector
    : 'Concept' '{' codeSelector (',' codeSelector)* '}' displayClause?
    ;

identifier
    : IDENTIFIER
    | DELIMITEDIDENTIFIER
    | QUOTEDIDENTIFIER
    ;

quantity
    : NUMBER unit?
    ;

unit
    : dateTimePrecision
    | pluralDateTimePrecision
    | STRING // UCUM syntax for units of measure
    ;

dateTimePrecision
    : 'year' | 'month' | 'week' | 'day' | 'hour' | 'minute' | 'second' | 'millisecond'
    ;

pluralDateTimePrecision
    : 'years' | 'months' | 'weeks' | 'days' | 'hours' | 'minutes' | 'seconds' | 'milliseconds'
    ;


/****************************************************************
    Lexical rules
*****************************************************************/

/*
NOTE: The goal of these rules in the grammar is to provide a date
token to the parser. As such it is not attempting to validate that
the date is a correct date, that task is for the parser or interpreter.
*/

DATE
    : '@' DATEFORMAT
    ;

DATETIME
    : '@' DATEFORMAT 'T' TIMEFORMAT? TIMEZONEOFFSETFORMAT?
    ;

TIME
    : '@' 'T' TIMEFORMAT
    ;

fragment DATEFORMAT
    : [0-9][0-9][0-9][0-9] ('-'[0-9][0-9] ('-'[0-9][0-9])?)?
    ;

fragment TIMEFORMAT
    : [0-9][0-9] (':'[0-9][0-9] (':'[0-9][0-9] ('.'[0-9]+)?)?)?
    ;

fragment TIMEZONEOFFSETFORMAT
    : ('Z' | ('+' | '-') [0-9][0-9]':'[0-9][0-9])
    ;

IDENTIFIER
    : ([A-Za-z] | '_')([A-Za-z0-9] | '_')*
    ;

DELIMITEDIDENTIFIER
    : '`' (ESC | .)*? '`'
    ;

QUOTEDIDENTIFIER
    : '"' (ESC | .)*? '"'
    ;

STRING
    : '\'' (ESC | .)*? '\''
    ;

// Also allows leading zeroes now (just like CQL and XSD)
NUMBER
    : ('+' | '-')?[0-9]+('.' [0-9]+)?
    ;

LONGNUMBER
    : ('+' | '-')?[0-9]+'L'
    ;

// Pipe whitespace to the HIDDEN channel to support retrieving source text through the parser.
WS
    : [ \r\n\t]+ -> channel(HIDDEN)
    ;

COMMENT
    : '/*' .*? '*/' -> channel(HIDDEN)
    ;

LINE_COMMENT
    : '//' ~[\r\n]* -> channel(HIDDEN)
    ;

fragment ESC
    : '\\' ([`'"\\/fnrt] | UNICODE)    // allow \`, \', \", \\, \/, \f, etc. and \uXXX
    ;

fragment UNICODE
    : 'u' HEX HEX HEX HEX
    ;

fragment HEX
    : [0-9a-fA-F]
    ;