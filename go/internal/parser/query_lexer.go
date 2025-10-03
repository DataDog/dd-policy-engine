// Code generated from ./internal/parser/QueryLexer.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type QueryLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var QueryLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func querylexerLexerInit() {
	staticData := &QueryLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE", "COMPARISON",
	}
	staticData.LiteralNames = []string{
		"", "'('", "')'", "':'", "'*'", "'\"'", "", "", "", "'<'", "'>'", "'<='",
		"'>='",
	}
	staticData.SymbolicNames = []string{
		"", "LPAREN", "RPAREN", "COLON", "STAR", "DQUOTE", "AND", "OR", "NOT",
		"LT", "GT", "LT_EQ", "GT_EQ", "TERM", "TERM_PREFIX", "TERM_SUFFIX",
		"TERM_CONTAINS", "PHRASE", "WS",
	}
	staticData.RuleNames = []string{
		"LPAREN", "RPAREN", "COLON", "STAR", "DQUOTE", "AND", "OR", "NOT", "LT",
		"GT", "LT_EQ", "GT_EQ", "ESC_CHAR", "STD_CHAR", "TERM", "TERM_START_CHAR",
		"TERM_CHAR", "TERM_PREFIX", "TERM_SUFFIX", "TERM_CONTAINS", "PHRASE",
		"WS", "NUMERIC_TERM", "DIGIT", "NUMBER", "COMPARISON_TERM",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 18, 192, 6, -1, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3,
		7, 3, 2, 4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9,
		7, 9, 2, 10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7,
		14, 2, 15, 7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19,
		2, 20, 7, 20, 2, 21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2,
		25, 7, 25, 1, 0, 1, 0, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 1, 3, 1, 4, 1, 4,
		1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 3, 5, 70, 8, 5, 1, 6, 1, 6, 1, 6, 1, 6, 3,
		6, 76, 8, 6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 3, 7, 83, 8, 7, 1, 8, 1, 8,
		1, 8, 1, 8, 1, 9, 1, 9, 1, 9, 1, 9, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10,
		1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 1, 12, 1, 13, 1, 13, 1,
		14, 1, 14, 5, 14, 110, 8, 14, 10, 14, 12, 14, 113, 9, 14, 1, 15, 1, 15,
		3, 15, 117, 8, 15, 1, 16, 1, 16, 3, 16, 121, 8, 16, 1, 17, 1, 17, 5, 17,
		125, 8, 17, 10, 17, 12, 17, 128, 9, 17, 1, 17, 1, 17, 1, 18, 1, 18, 4,
		18, 134, 8, 18, 11, 18, 12, 18, 135, 1, 19, 1, 19, 4, 19, 140, 8, 19, 11,
		19, 12, 19, 141, 1, 19, 1, 19, 1, 20, 1, 20, 1, 20, 5, 20, 149, 8, 20,
		10, 20, 12, 20, 152, 9, 20, 1, 20, 1, 20, 1, 21, 4, 21, 157, 8, 21, 11,
		21, 12, 21, 158, 1, 21, 1, 21, 1, 22, 1, 22, 1, 22, 3, 22, 166, 8, 22,
		1, 23, 1, 23, 1, 24, 3, 24, 171, 8, 24, 1, 24, 4, 24, 174, 8, 24, 11, 24,
		12, 24, 175, 1, 24, 1, 24, 4, 24, 180, 8, 24, 11, 24, 12, 24, 181, 3, 24,
		184, 8, 24, 1, 25, 1, 25, 3, 25, 188, 8, 25, 1, 25, 1, 25, 1, 25, 1, 150,
		0, 26, 2, 1, 4, 2, 6, 3, 8, 4, 10, 5, 12, 6, 14, 7, 16, 8, 18, 9, 20, 10,
		22, 11, 24, 12, 26, 0, 28, 0, 30, 13, 32, 0, 34, 0, 36, 14, 38, 15, 40,
		16, 42, 17, 44, 18, 46, 0, 48, 0, 50, 0, 52, 0, 2, 0, 1, 5, 2, 0, 33, 33,
		45, 45, 11, 0, 9, 10, 13, 13, 32, 34, 40, 42, 45, 45, 58, 58, 60, 60, 62,
		63, 91, 93, 123, 125, 12288, 12288, 2, 0, 45, 45, 60, 62, 1, 0, 34, 34,
		3, 0, 9, 10, 13, 13, 32, 32, 201, 0, 2, 1, 0, 0, 0, 0, 4, 1, 0, 0, 0, 0,
		6, 1, 0, 0, 0, 0, 8, 1, 0, 0, 0, 0, 10, 1, 0, 0, 0, 0, 12, 1, 0, 0, 0,
		0, 14, 1, 0, 0, 0, 0, 16, 1, 0, 0, 0, 0, 18, 1, 0, 0, 0, 0, 20, 1, 0, 0,
		0, 0, 22, 1, 0, 0, 0, 0, 24, 1, 0, 0, 0, 0, 30, 1, 0, 0, 0, 0, 36, 1, 0,
		0, 0, 0, 38, 1, 0, 0, 0, 0, 40, 1, 0, 0, 0, 0, 42, 1, 0, 0, 0, 0, 44, 1,
		0, 0, 0, 1, 52, 1, 0, 0, 0, 2, 54, 1, 0, 0, 0, 4, 56, 1, 0, 0, 0, 6, 58,
		1, 0, 0, 0, 8, 60, 1, 0, 0, 0, 10, 62, 1, 0, 0, 0, 12, 69, 1, 0, 0, 0,
		14, 75, 1, 0, 0, 0, 16, 82, 1, 0, 0, 0, 18, 84, 1, 0, 0, 0, 20, 88, 1,
		0, 0, 0, 22, 92, 1, 0, 0, 0, 24, 97, 1, 0, 0, 0, 26, 102, 1, 0, 0, 0, 28,
		105, 1, 0, 0, 0, 30, 107, 1, 0, 0, 0, 32, 116, 1, 0, 0, 0, 34, 120, 1,
		0, 0, 0, 36, 122, 1, 0, 0, 0, 38, 131, 1, 0, 0, 0, 40, 137, 1, 0, 0, 0,
		42, 145, 1, 0, 0, 0, 44, 156, 1, 0, 0, 0, 46, 162, 1, 0, 0, 0, 48, 167,
		1, 0, 0, 0, 50, 170, 1, 0, 0, 0, 52, 187, 1, 0, 0, 0, 54, 55, 5, 40, 0,
		0, 55, 3, 1, 0, 0, 0, 56, 57, 5, 41, 0, 0, 57, 5, 1, 0, 0, 0, 58, 59, 5,
		58, 0, 0, 59, 7, 1, 0, 0, 0, 60, 61, 5, 42, 0, 0, 61, 9, 1, 0, 0, 0, 62,
		63, 5, 34, 0, 0, 63, 11, 1, 0, 0, 0, 64, 65, 5, 65, 0, 0, 65, 66, 5, 78,
		0, 0, 66, 70, 5, 68, 0, 0, 67, 68, 5, 38, 0, 0, 68, 70, 5, 38, 0, 0, 69,
		64, 1, 0, 0, 0, 69, 67, 1, 0, 0, 0, 70, 13, 1, 0, 0, 0, 71, 72, 5, 79,
		0, 0, 72, 76, 5, 82, 0, 0, 73, 74, 5, 124, 0, 0, 74, 76, 5, 124, 0, 0,
		75, 71, 1, 0, 0, 0, 75, 73, 1, 0, 0, 0, 76, 15, 1, 0, 0, 0, 77, 78, 5,
		78, 0, 0, 78, 79, 5, 79, 0, 0, 79, 83, 5, 84, 0, 0, 80, 81, 7, 0, 0, 0,
		81, 83, 4, 7, 0, 0, 82, 77, 1, 0, 0, 0, 82, 80, 1, 0, 0, 0, 83, 17, 1,
		0, 0, 0, 84, 85, 5, 60, 0, 0, 85, 86, 1, 0, 0, 0, 86, 87, 6, 8, 0, 0, 87,
		19, 1, 0, 0, 0, 88, 89, 5, 62, 0, 0, 89, 90, 1, 0, 0, 0, 90, 91, 6, 9,
		0, 0, 91, 21, 1, 0, 0, 0, 92, 93, 5, 60, 0, 0, 93, 94, 5, 61, 0, 0, 94,
		95, 1, 0, 0, 0, 95, 96, 6, 10, 0, 0, 96, 23, 1, 0, 0, 0, 97, 98, 5, 62,
		0, 0, 98, 99, 5, 61, 0, 0, 99, 100, 1, 0, 0, 0, 100, 101, 6, 11, 0, 0,
		101, 25, 1, 0, 0, 0, 102, 103, 5, 92, 0, 0, 103, 104, 9, 0, 0, 0, 104,
		27, 1, 0, 0, 0, 105, 106, 8, 1, 0, 0, 106, 29, 1, 0, 0, 0, 107, 111, 3,
		32, 15, 0, 108, 110, 3, 34, 16, 0, 109, 108, 1, 0, 0, 0, 110, 113, 1, 0,
		0, 0, 111, 109, 1, 0, 0, 0, 111, 112, 1, 0, 0, 0, 112, 31, 1, 0, 0, 0,
		113, 111, 1, 0, 0, 0, 114, 117, 3, 28, 13, 0, 115, 117, 3, 26, 12, 0, 116,
		114, 1, 0, 0, 0, 116, 115, 1, 0, 0, 0, 117, 33, 1, 0, 0, 0, 118, 121, 3,
		32, 15, 0, 119, 121, 7, 2, 0, 0, 120, 118, 1, 0, 0, 0, 120, 119, 1, 0,
		0, 0, 121, 35, 1, 0, 0, 0, 122, 126, 3, 32, 15, 0, 123, 125, 3, 34, 16,
		0, 124, 123, 1, 0, 0, 0, 125, 128, 1, 0, 0, 0, 126, 124, 1, 0, 0, 0, 126,
		127, 1, 0, 0, 0, 127, 129, 1, 0, 0, 0, 128, 126, 1, 0, 0, 0, 129, 130,
		3, 8, 3, 0, 130, 37, 1, 0, 0, 0, 131, 133, 3, 8, 3, 0, 132, 134, 3, 34,
		16, 0, 133, 132, 1, 0, 0, 0, 134, 135, 1, 0, 0, 0, 135, 133, 1, 0, 0, 0,
		135, 136, 1, 0, 0, 0, 136, 39, 1, 0, 0, 0, 137, 139, 3, 8, 3, 0, 138, 140,
		3, 34, 16, 0, 139, 138, 1, 0, 0, 0, 140, 141, 1, 0, 0, 0, 141, 139, 1,
		0, 0, 0, 141, 142, 1, 0, 0, 0, 142, 143, 1, 0, 0, 0, 143, 144, 3, 8, 3,
		0, 144, 41, 1, 0, 0, 0, 145, 150, 3, 10, 4, 0, 146, 149, 3, 26, 12, 0,
		147, 149, 8, 3, 0, 0, 148, 146, 1, 0, 0, 0, 148, 147, 1, 0, 0, 0, 149,
		152, 1, 0, 0, 0, 150, 151, 1, 0, 0, 0, 150, 148, 1, 0, 0, 0, 151, 153,
		1, 0, 0, 0, 152, 150, 1, 0, 0, 0, 153, 154, 3, 10, 4, 0, 154, 43, 1, 0,
		0, 0, 155, 157, 7, 4, 0, 0, 156, 155, 1, 0, 0, 0, 157, 158, 1, 0, 0, 0,
		158, 156, 1, 0, 0, 0, 158, 159, 1, 0, 0, 0, 159, 160, 1, 0, 0, 0, 160,
		161, 6, 21, 1, 0, 161, 45, 1, 0, 0, 0, 162, 165, 3, 50, 24, 0, 163, 164,
		5, 69, 0, 0, 164, 166, 3, 50, 24, 0, 165, 163, 1, 0, 0, 0, 165, 166, 1,
		0, 0, 0, 166, 47, 1, 0, 0, 0, 167, 168, 2, 48, 57, 0, 168, 49, 1, 0, 0,
		0, 169, 171, 5, 45, 0, 0, 170, 169, 1, 0, 0, 0, 170, 171, 1, 0, 0, 0, 171,
		173, 1, 0, 0, 0, 172, 174, 3, 48, 23, 0, 173, 172, 1, 0, 0, 0, 174, 175,
		1, 0, 0, 0, 175, 173, 1, 0, 0, 0, 175, 176, 1, 0, 0, 0, 176, 183, 1, 0,
		0, 0, 177, 179, 5, 46, 0, 0, 178, 180, 3, 48, 23, 0, 179, 178, 1, 0, 0,
		0, 180, 181, 1, 0, 0, 0, 181, 179, 1, 0, 0, 0, 181, 182, 1, 0, 0, 0, 182,
		184, 1, 0, 0, 0, 183, 177, 1, 0, 0, 0, 183, 184, 1, 0, 0, 0, 184, 51, 1,
		0, 0, 0, 185, 188, 3, 46, 22, 0, 186, 188, 3, 30, 14, 0, 187, 185, 1, 0,
		0, 0, 187, 186, 1, 0, 0, 0, 188, 189, 1, 0, 0, 0, 189, 190, 6, 25, 2, 0,
		190, 191, 6, 25, 3, 0, 191, 53, 1, 0, 0, 0, 20, 0, 1, 69, 75, 82, 111,
		116, 120, 126, 135, 141, 148, 150, 158, 165, 170, 175, 181, 183, 187, 4,
		5, 1, 0, 6, 0, 0, 7, 13, 0, 4, 0, 0,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// QueryLexerInit initializes any static state used to implement QueryLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewQueryLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func QueryLexerInit() {
	staticData := &QueryLexerLexerStaticData
	staticData.once.Do(querylexerLexerInit)
}

// NewQueryLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewQueryLexer(input antlr.CharStream) *QueryLexer {
	QueryLexerInit()
	l := new(QueryLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &QueryLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "QueryLexer.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// QueryLexer tokens.
const (
	QueryLexerLPAREN        = 1
	QueryLexerRPAREN        = 2
	QueryLexerCOLON         = 3
	QueryLexerSTAR          = 4
	QueryLexerDQUOTE        = 5
	QueryLexerAND           = 6
	QueryLexerOR            = 7
	QueryLexerNOT           = 8
	QueryLexerLT            = 9
	QueryLexerGT            = 10
	QueryLexerLT_EQ         = 11
	QueryLexerGT_EQ         = 12
	QueryLexerTERM          = 13
	QueryLexerTERM_PREFIX   = 14
	QueryLexerTERM_SUFFIX   = 15
	QueryLexerTERM_CONTAINS = 16
	QueryLexerPHRASE        = 17
	QueryLexerWS            = 18
)

// QueryLexerCOMPARISON is the QueryLexer mode.
const QueryLexerCOMPARISON = 1

func (l *QueryLexer) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 7:
		return l.NOT_Sempred(localctx, predIndex)

	default:
		panic("No registered predicate for: " + fmt.Sprint(ruleIndex))
	}
}

func (p *QueryLexer) NOT_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.GetInputStream().LA(1) != ' ' && p.GetInputStream().LA(1) != '\t'

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
