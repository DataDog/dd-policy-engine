parser grammar QueryParser;

options { tokenVocab=QueryLexer; }

topLevelQuery
  : orExpr EOF
  ;

orExpr
    : andExpr (OR andExpr)*
    ;

andExpr
    : modClause (AND modClause)*
    | modClause (modClause)*
    ;

modClause
  : modifiers clause
  ;

modifiers
  : NOT
  |
  ;

clause
  : field (simpleValue /*| advancedValue*/)
  | LPAREN orExpr RPAREN
  ;

field
  : TERM COLON
  ;
/*
advancedValue
  : comparison
  ;
 */

simpleValue
  : normal
  // | quoted
  | prefix
  | suffix
  | contains
  ;

normal
  : TERM
  ;

/*
quoted
  : PHRASE
  ;
*/

prefix
  : TERM_PREFIX
  ;

suffix
  : TERM_SUFFIX
  ;

contains
  : TERM_CONTAINS
  ;

/*
comparison
  : operator TERM
  ;

operator
  : LT
  | GT
  | LT_EQ
  | GT_EQ
  ;
  */
