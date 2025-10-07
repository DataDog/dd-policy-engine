lexer grammar QueryLexer;

LPAREN
  : '('
  ;

RPAREN
  : ')'
  ;

COLON
  : ':'
  ;

STAR
  : '*'
  ;

DQUOTE
  : '"'
  ;

AND
  : 'AND'
  | '&&'
  ;

OR
  : 'OR'
  | '||'
  ;

NOT
  : 'NOT'
  | ('-' | '!') { p.GetInputStream().LA(1) != ' ' && p.GetInputStream().LA(1) != '\t' }?
  ;

LT
  : '<' -> pushMode(COMPARISON)
  ;

GT
  : '>' -> pushMode(COMPARISON)
  ;

LT_EQ
  : '<='  -> pushMode(COMPARISON)
  ;

GT_EQ
  : '>='  -> pushMode(COMPARISON)
  ;

// All characters are technically valid when they're escaped.
fragment ESC_CHAR
  : '\\' .
  ;

fragment STD_CHAR
  : ~(' ' | '\t' | '\n' | '\r' | '\u3000' | '-' | '>' | '<' | '!' | '(' | ')' | '{' | '}' | '[' | ']' | '"' | '*' | '?' | ':' | '\\' | '|')
  ;

TERM
  // Require at least one TERM_START_CHAR so that TERM cannot match empty strings.
  : TERM_START_CHAR (TERM_CHAR)*
  ;

fragment TERM_START_CHAR
  : STD_CHAR
  | ESC_CHAR
  ;

fragment TERM_CHAR
  : (TERM_START_CHAR | '-' | '=' | '>' | '<' )
  ;

TERM_PREFIX
  : TERM_START_CHAR (TERM_CHAR)* STAR
  ;

TERM_SUFFIX
  : STAR (TERM_CHAR)+
  ;

TERM_CONTAINS
  : STAR (TERM_CHAR)+ STAR
  ;

PHRASE
  : DQUOTE (ESC_CHAR|~('"'))*? DQUOTE
  ;

WS
   : [ \r\n\t]+ -> skip
   ;

/* Comparison submode to properly recognize numbers with/without decimals and E notation */
mode COMPARISON;

fragment NUMERIC_TERM
  : NUMBER ('E' NUMBER)?
  ;

fragment DIGIT : ('0'..'9') ;

fragment NUMBER
  : '-'? DIGIT+ ('.' DIGIT+)?
  ;

COMPARISON_TERM
  : (NUMERIC_TERM
  | TERM) -> type(TERM), popMode
  ;
