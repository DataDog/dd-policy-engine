package parser

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TokenType int

const (
	tokenError TokenType = iota
	tokenKey
	tokenValue
	tokenAnd
	tokenOr
	tokenLParenthesis
	tokenRParenthesis
	tokenEOF
)

type Token struct {
	Type   TokenType
	Value  string
	Start  int
	Length int
}

func (t Token) String() string {
	switch t.Type {
	case tokenEOF:
		return "EOF"
	case tokenError:
		return fmt.Sprintf("error: %v. Pos: %d", t.Value, t.Start)
	case tokenAnd:
		return "AND"
	case tokenOr:
		return "OR"
	}

	if len(t.Value) > 10 {
		return fmt.Sprintf("%.10q...", t.Value)
	}
	return fmt.Sprintf("%q", t.Value)
}

type stateFn func(*lexer) stateFn

type lexer struct {
	input  string
	start  int
	end    int
	pos    int
	width  int
	tokens chan Token
}

func lex(input string) (*lexer, chan Token) {
	l := &lexer{
		input:  input,
		start:  0,
		end:    len(input),
		tokens: make(chan Token),
	}

	go l.run()
	return l, l.tokens
}

const EOF = -1

func (l *lexer) next() (rune rune) {
	if l.pos >= l.end {
		return EOF
	}

	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return rune
}

// NOTE: ignore the next char
func (l *lexer) ignore() {
	l.advance(1)
}

func (l *lexer) consume() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

func (l *lexer) advance(n int) {
	l.pos += n
	l.start = l.pos
}

func (l *lexer) emit(token Token) {
	l.tokens <- token
}

func (l *lexer) run() {
	for state := termState; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

func termState(l *lexer) stateFn {
	r := l.next()
	for r != EOF {
		switch {
		case unicode.IsLetter(r):
			l.backup()
			return keyState
		case unicode.IsSpace(r):
			l.consume()
		default:
			l.backup()
			l.emit(Token{Type: tokenError, Value: "unexpected char", Start: l.pos})
			return nil
		}

		r = l.next()
	}

	l.emit(Token{Type: tokenEOF})
	return nil
}

func keyState(l *lexer) stateFn {
	r := l.next()
	for r != EOF {
		switch {
		case r == ':':
			l.backup()
			key := l.input[l.start:l.pos]
			l.consume()
			l.emit(Token{Type: tokenKey, Value: key})
			l.ignore() //< ignore ':'
			return valueState
		case unicode.IsSpace(r):
			l.backup()
			l.emit(Token{Type: tokenError, Value: "unexpected character", Start: l.pos})
			return nil
		}
		r = l.next()
	}

	l.emit(Token{Type: tokenError, Value: "expected ':' but got EOF", Start: l.pos})
	return nil
}

func valueState(l *lexer) stateFn {
	r := l.next()

	// A valid value has at least one valid character
	if r == EOF {
		l.emit(Token{Type: tokenError, Value: "expected a character but got EOF", Start: l.pos})
		return nil
	} else if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
		l.emit(Token{Type: tokenError, Value: "expected a character but got non supported character", Start: l.pos})
		return nil
	}

	for r != EOF {
		if unicode.IsSpace(r) {
			l.backup()
			value := l.input[l.start:l.pos]
			l.consume()
			l.emit(Token{Type: tokenValue, Value: value})
			return conjonctionTerm
		}
		r = l.next()
	}

	l.emit(Token{Type: tokenValue, Value: l.input[l.start:l.pos]})
	l.emit(Token{Type: tokenEOF})
	return nil
}

func conjonctionTerm(l *lexer) stateFn {
	r := l.peek()
	for r != EOF {
		switch {
		case r == 'a' || r == 'A':
			return maybeAndTerm
		case r == 'o' || r == 'O':
			return maybeOrTerm
		case unicode.IsLetter(r):
			// Implicit AND
			l.emit(Token{Type: tokenAnd})
			return keyState
		case unicode.IsSpace(r):
			// TODO: Do i need this consume?
			l.consume()
		default:
			l.emit(Token{Type: tokenError, Value: "Nope", Start: l.pos})
			return nil
		}
		l.ignore()
		r = l.peek()
	}

	l.emit(Token{Type: tokenEOF})
	return nil
}

func maybeAndTerm(l *lexer) stateFn {
	const andTerm = "and " //< need at least a space after `and`
	if l.pos+len(andTerm) >= l.end {
		l.emit(Token{Type: tokenAnd})
		return keyState
	}

	term := strings.ToLower(l.input[l.pos : l.pos+len(andTerm)])
	if term != andTerm {
		l.emit(Token{Type: tokenAnd})
		return keyState
	}

	l.advance(len(andTerm))
	l.emit(Token{Type: tokenAnd})
	return termState
}

func maybeOrTerm(l *lexer) stateFn {
	const orTerm = "or " //< need at least a space after `or`
	if l.pos+len(orTerm) >= l.end {
		l.emit(Token{Type: tokenAnd})
		return keyState
	}

	term := strings.ToLower(l.input[l.pos : l.pos+len(orTerm)])
	if term != orTerm {
		l.emit(Token{Type: tokenAnd})
		return keyState
	}

	l.advance(len(orTerm))
	l.emit(Token{Type: tokenOr})
	return termState
}
