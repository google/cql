// Code generated from Cvlt.g4 by ANTLR 4.13.1. DO NOT EDIT.

package cvlt // Cvlt
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type CvltParser struct {
	*antlr.BaseParser
}

var CvltParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func cvltParserInit() {
	staticData := &CvltParserStaticData
	staticData.LiteralNames = []string{
		"", "'.'", "'List'", "'<'", "'>'", "'Interval'", "'Tuple'", "'{'", "','",
		"'}'", "'Choice'", "':'", "'true'", "'false'", "'null'", "'['", "'('",
		"']'", "')'", "'display'", "'Code'", "'from'", "'Concept'", "'year'",
		"'month'", "'week'", "'day'", "'hour'", "'minute'", "'second'", "'millisecond'",
		"'years'", "'months'", "'weeks'", "'days'", "'hours'", "'minutes'",
		"'seconds'", "'milliseconds'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "DATE", "DATETIME", "TIME", "IDENTIFIER", "DELIMITEDIDENTIFIER",
		"QUOTEDIDENTIFIER", "STRING", "NUMBER", "LONGNUMBER", "WS", "COMMENT",
		"LINE_COMMENT",
	}
	staticData.RuleNames = []string{
		"typeSpecifier", "namedTypeSpecifier", "modelIdentifier", "listTypeSpecifier",
		"intervalTypeSpecifier", "tupleTypeSpecifier", "tupleElementDefinition",
		"choiceTypeSpecifier", "term", "ratio", "literal", "intervalSelector",
		"tupleSelector", "tupleElementSelector", "instanceSelector", "instanceElementSelector",
		"listSelector", "displayClause", "codeSelector", "conceptSelector",
		"identifier", "quantity", "unit", "dateTimePrecision", "pluralDateTimePrecision",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 50, 240, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2,
		21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 1, 0, 1, 0, 1, 0,
		1, 0, 1, 0, 3, 0, 56, 8, 0, 1, 1, 1, 1, 1, 1, 5, 1, 61, 8, 1, 10, 1, 12,
		1, 64, 9, 1, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 4,
		1, 4, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 5, 5, 85, 8, 5, 10,
		5, 12, 5, 88, 9, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 5, 7, 100, 8, 7, 10, 7, 12, 7, 103, 9, 7, 1, 7, 1, 7, 1, 8, 1,
		8, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8, 3, 8, 114, 8, 8, 1, 9, 1, 9, 1, 9, 1,
		9, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10,
		3, 10, 130, 8, 10, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1,
		12, 3, 12, 140, 8, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 5, 12, 147, 8,
		12, 10, 12, 12, 12, 150, 9, 12, 3, 12, 152, 8, 12, 1, 12, 1, 12, 1, 13,
		1, 13, 1, 13, 1, 13, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 5, 14, 166,
		8, 14, 10, 14, 12, 14, 169, 9, 14, 3, 14, 171, 8, 14, 1, 14, 1, 14, 1,
		15, 1, 15, 1, 15, 1, 15, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 3, 16, 184,
		8, 16, 3, 16, 186, 8, 16, 1, 16, 1, 16, 1, 16, 1, 16, 5, 16, 192, 8, 16,
		10, 16, 12, 16, 195, 9, 16, 3, 16, 197, 8, 16, 1, 16, 1, 16, 1, 17, 1,
		17, 1, 17, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 3, 18, 209, 8, 18, 1, 19,
		1, 19, 1, 19, 1, 19, 1, 19, 5, 19, 216, 8, 19, 10, 19, 12, 19, 219, 9,
		19, 1, 19, 1, 19, 3, 19, 223, 8, 19, 1, 20, 1, 20, 1, 21, 1, 21, 3, 21,
		229, 8, 21, 1, 22, 1, 22, 1, 22, 3, 22, 234, 8, 22, 1, 23, 1, 23, 1, 24,
		1, 24, 1, 24, 0, 0, 25, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24,
		26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 0, 6, 1, 0, 12, 13, 1,
		0, 15, 16, 1, 0, 17, 18, 1, 0, 42, 44, 1, 0, 23, 30, 1, 0, 31, 38, 251,
		0, 55, 1, 0, 0, 0, 2, 62, 1, 0, 0, 0, 4, 67, 1, 0, 0, 0, 6, 69, 1, 0, 0,
		0, 8, 74, 1, 0, 0, 0, 10, 79, 1, 0, 0, 0, 12, 91, 1, 0, 0, 0, 14, 94, 1,
		0, 0, 0, 16, 113, 1, 0, 0, 0, 18, 115, 1, 0, 0, 0, 20, 129, 1, 0, 0, 0,
		22, 131, 1, 0, 0, 0, 24, 139, 1, 0, 0, 0, 26, 155, 1, 0, 0, 0, 28, 159,
		1, 0, 0, 0, 30, 174, 1, 0, 0, 0, 32, 185, 1, 0, 0, 0, 34, 200, 1, 0, 0,
		0, 36, 203, 1, 0, 0, 0, 38, 210, 1, 0, 0, 0, 40, 224, 1, 0, 0, 0, 42, 226,
		1, 0, 0, 0, 44, 233, 1, 0, 0, 0, 46, 235, 1, 0, 0, 0, 48, 237, 1, 0, 0,
		0, 50, 56, 3, 2, 1, 0, 51, 56, 3, 6, 3, 0, 52, 56, 3, 8, 4, 0, 53, 56,
		3, 10, 5, 0, 54, 56, 3, 14, 7, 0, 55, 50, 1, 0, 0, 0, 55, 51, 1, 0, 0,
		0, 55, 52, 1, 0, 0, 0, 55, 53, 1, 0, 0, 0, 55, 54, 1, 0, 0, 0, 56, 1, 1,
		0, 0, 0, 57, 58, 3, 40, 20, 0, 58, 59, 5, 1, 0, 0, 59, 61, 1, 0, 0, 0,
		60, 57, 1, 0, 0, 0, 61, 64, 1, 0, 0, 0, 62, 60, 1, 0, 0, 0, 62, 63, 1,
		0, 0, 0, 63, 65, 1, 0, 0, 0, 64, 62, 1, 0, 0, 0, 65, 66, 3, 40, 20, 0,
		66, 3, 1, 0, 0, 0, 67, 68, 3, 40, 20, 0, 68, 5, 1, 0, 0, 0, 69, 70, 5,
		2, 0, 0, 70, 71, 5, 3, 0, 0, 71, 72, 3, 0, 0, 0, 72, 73, 5, 4, 0, 0, 73,
		7, 1, 0, 0, 0, 74, 75, 5, 5, 0, 0, 75, 76, 5, 3, 0, 0, 76, 77, 3, 0, 0,
		0, 77, 78, 5, 4, 0, 0, 78, 9, 1, 0, 0, 0, 79, 80, 5, 6, 0, 0, 80, 81, 5,
		7, 0, 0, 81, 86, 3, 12, 6, 0, 82, 83, 5, 8, 0, 0, 83, 85, 3, 12, 6, 0,
		84, 82, 1, 0, 0, 0, 85, 88, 1, 0, 0, 0, 86, 84, 1, 0, 0, 0, 86, 87, 1,
		0, 0, 0, 87, 89, 1, 0, 0, 0, 88, 86, 1, 0, 0, 0, 89, 90, 5, 9, 0, 0, 90,
		11, 1, 0, 0, 0, 91, 92, 3, 40, 20, 0, 92, 93, 3, 0, 0, 0, 93, 13, 1, 0,
		0, 0, 94, 95, 5, 10, 0, 0, 95, 96, 5, 3, 0, 0, 96, 101, 3, 0, 0, 0, 97,
		98, 5, 8, 0, 0, 98, 100, 3, 0, 0, 0, 99, 97, 1, 0, 0, 0, 100, 103, 1, 0,
		0, 0, 101, 99, 1, 0, 0, 0, 101, 102, 1, 0, 0, 0, 102, 104, 1, 0, 0, 0,
		103, 101, 1, 0, 0, 0, 104, 105, 5, 4, 0, 0, 105, 15, 1, 0, 0, 0, 106, 114,
		3, 20, 10, 0, 107, 114, 3, 22, 11, 0, 108, 114, 3, 24, 12, 0, 109, 114,
		3, 28, 14, 0, 110, 114, 3, 32, 16, 0, 111, 114, 3, 36, 18, 0, 112, 114,
		3, 38, 19, 0, 113, 106, 1, 0, 0, 0, 113, 107, 1, 0, 0, 0, 113, 108, 1,
		0, 0, 0, 113, 109, 1, 0, 0, 0, 113, 110, 1, 0, 0, 0, 113, 111, 1, 0, 0,
		0, 113, 112, 1, 0, 0, 0, 114, 17, 1, 0, 0, 0, 115, 116, 3, 42, 21, 0, 116,
		117, 5, 11, 0, 0, 117, 118, 3, 42, 21, 0, 118, 19, 1, 0, 0, 0, 119, 130,
		7, 0, 0, 0, 120, 130, 5, 14, 0, 0, 121, 130, 5, 45, 0, 0, 122, 130, 5,
		46, 0, 0, 123, 130, 5, 47, 0, 0, 124, 130, 5, 40, 0, 0, 125, 130, 5, 39,
		0, 0, 126, 130, 5, 41, 0, 0, 127, 130, 3, 42, 21, 0, 128, 130, 3, 18, 9,
		0, 129, 119, 1, 0, 0, 0, 129, 120, 1, 0, 0, 0, 129, 121, 1, 0, 0, 0, 129,
		122, 1, 0, 0, 0, 129, 123, 1, 0, 0, 0, 129, 124, 1, 0, 0, 0, 129, 125,
		1, 0, 0, 0, 129, 126, 1, 0, 0, 0, 129, 127, 1, 0, 0, 0, 129, 128, 1, 0,
		0, 0, 130, 21, 1, 0, 0, 0, 131, 132, 5, 5, 0, 0, 132, 133, 7, 1, 0, 0,
		133, 134, 3, 20, 10, 0, 134, 135, 5, 8, 0, 0, 135, 136, 3, 20, 10, 0, 136,
		137, 7, 2, 0, 0, 137, 23, 1, 0, 0, 0, 138, 140, 5, 6, 0, 0, 139, 138, 1,
		0, 0, 0, 139, 140, 1, 0, 0, 0, 140, 141, 1, 0, 0, 0, 141, 151, 5, 7, 0,
		0, 142, 152, 5, 11, 0, 0, 143, 148, 3, 26, 13, 0, 144, 145, 5, 8, 0, 0,
		145, 147, 3, 26, 13, 0, 146, 144, 1, 0, 0, 0, 147, 150, 1, 0, 0, 0, 148,
		146, 1, 0, 0, 0, 148, 149, 1, 0, 0, 0, 149, 152, 1, 0, 0, 0, 150, 148,
		1, 0, 0, 0, 151, 142, 1, 0, 0, 0, 151, 143, 1, 0, 0, 0, 152, 153, 1, 0,
		0, 0, 153, 154, 5, 9, 0, 0, 154, 25, 1, 0, 0, 0, 155, 156, 3, 40, 20, 0,
		156, 157, 5, 11, 0, 0, 157, 158, 3, 16, 8, 0, 158, 27, 1, 0, 0, 0, 159,
		160, 3, 40, 20, 0, 160, 170, 5, 7, 0, 0, 161, 171, 5, 11, 0, 0, 162, 167,
		3, 30, 15, 0, 163, 164, 5, 8, 0, 0, 164, 166, 3, 30, 15, 0, 165, 163, 1,
		0, 0, 0, 166, 169, 1, 0, 0, 0, 167, 165, 1, 0, 0, 0, 167, 168, 1, 0, 0,
		0, 168, 171, 1, 0, 0, 0, 169, 167, 1, 0, 0, 0, 170, 161, 1, 0, 0, 0, 170,
		162, 1, 0, 0, 0, 171, 172, 1, 0, 0, 0, 172, 173, 5, 9, 0, 0, 173, 29, 1,
		0, 0, 0, 174, 175, 3, 40, 20, 0, 175, 176, 5, 11, 0, 0, 176, 177, 3, 16,
		8, 0, 177, 31, 1, 0, 0, 0, 178, 183, 5, 2, 0, 0, 179, 180, 5, 3, 0, 0,
		180, 181, 3, 0, 0, 0, 181, 182, 5, 4, 0, 0, 182, 184, 1, 0, 0, 0, 183,
		179, 1, 0, 0, 0, 183, 184, 1, 0, 0, 0, 184, 186, 1, 0, 0, 0, 185, 178,
		1, 0, 0, 0, 185, 186, 1, 0, 0, 0, 186, 187, 1, 0, 0, 0, 187, 196, 5, 7,
		0, 0, 188, 193, 3, 16, 8, 0, 189, 190, 5, 8, 0, 0, 190, 192, 3, 16, 8,
		0, 191, 189, 1, 0, 0, 0, 192, 195, 1, 0, 0, 0, 193, 191, 1, 0, 0, 0, 193,
		194, 1, 0, 0, 0, 194, 197, 1, 0, 0, 0, 195, 193, 1, 0, 0, 0, 196, 188,
		1, 0, 0, 0, 196, 197, 1, 0, 0, 0, 197, 198, 1, 0, 0, 0, 198, 199, 5, 9,
		0, 0, 199, 33, 1, 0, 0, 0, 200, 201, 5, 19, 0, 0, 201, 202, 5, 45, 0, 0,
		202, 35, 1, 0, 0, 0, 203, 204, 5, 20, 0, 0, 204, 205, 5, 45, 0, 0, 205,
		206, 5, 21, 0, 0, 206, 208, 3, 40, 20, 0, 207, 209, 3, 34, 17, 0, 208,
		207, 1, 0, 0, 0, 208, 209, 1, 0, 0, 0, 209, 37, 1, 0, 0, 0, 210, 211, 5,
		22, 0, 0, 211, 212, 5, 7, 0, 0, 212, 217, 3, 36, 18, 0, 213, 214, 5, 8,
		0, 0, 214, 216, 3, 36, 18, 0, 215, 213, 1, 0, 0, 0, 216, 219, 1, 0, 0,
		0, 217, 215, 1, 0, 0, 0, 217, 218, 1, 0, 0, 0, 218, 220, 1, 0, 0, 0, 219,
		217, 1, 0, 0, 0, 220, 222, 5, 9, 0, 0, 221, 223, 3, 34, 17, 0, 222, 221,
		1, 0, 0, 0, 222, 223, 1, 0, 0, 0, 223, 39, 1, 0, 0, 0, 224, 225, 7, 3,
		0, 0, 225, 41, 1, 0, 0, 0, 226, 228, 5, 46, 0, 0, 227, 229, 3, 44, 22,
		0, 228, 227, 1, 0, 0, 0, 228, 229, 1, 0, 0, 0, 229, 43, 1, 0, 0, 0, 230,
		234, 3, 46, 23, 0, 231, 234, 3, 48, 24, 0, 232, 234, 5, 45, 0, 0, 233,
		230, 1, 0, 0, 0, 233, 231, 1, 0, 0, 0, 233, 232, 1, 0, 0, 0, 234, 45, 1,
		0, 0, 0, 235, 236, 7, 4, 0, 0, 236, 47, 1, 0, 0, 0, 237, 238, 7, 5, 0,
		0, 238, 49, 1, 0, 0, 0, 20, 55, 62, 86, 101, 113, 129, 139, 148, 151, 167,
		170, 183, 185, 193, 196, 208, 217, 222, 228, 233,
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

// CvltParserInit initializes any static state used to implement CvltParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewCvltParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func CvltParserInit() {
	staticData := &CvltParserStaticData
	staticData.once.Do(cvltParserInit)
}

// NewCvltParser produces a new parser instance for the optional input antlr.TokenStream.
func NewCvltParser(input antlr.TokenStream) *CvltParser {
	CvltParserInit()
	this := new(CvltParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &CvltParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "Cvlt.g4"

	return this
}

// CvltParser tokens.
const (
	CvltParserEOF                 = antlr.TokenEOF
	CvltParserT__0                = 1
	CvltParserT__1                = 2
	CvltParserT__2                = 3
	CvltParserT__3                = 4
	CvltParserT__4                = 5
	CvltParserT__5                = 6
	CvltParserT__6                = 7
	CvltParserT__7                = 8
	CvltParserT__8                = 9
	CvltParserT__9                = 10
	CvltParserT__10               = 11
	CvltParserT__11               = 12
	CvltParserT__12               = 13
	CvltParserT__13               = 14
	CvltParserT__14               = 15
	CvltParserT__15               = 16
	CvltParserT__16               = 17
	CvltParserT__17               = 18
	CvltParserT__18               = 19
	CvltParserT__19               = 20
	CvltParserT__20               = 21
	CvltParserT__21               = 22
	CvltParserT__22               = 23
	CvltParserT__23               = 24
	CvltParserT__24               = 25
	CvltParserT__25               = 26
	CvltParserT__26               = 27
	CvltParserT__27               = 28
	CvltParserT__28               = 29
	CvltParserT__29               = 30
	CvltParserT__30               = 31
	CvltParserT__31               = 32
	CvltParserT__32               = 33
	CvltParserT__33               = 34
	CvltParserT__34               = 35
	CvltParserT__35               = 36
	CvltParserT__36               = 37
	CvltParserT__37               = 38
	CvltParserDATE                = 39
	CvltParserDATETIME            = 40
	CvltParserTIME                = 41
	CvltParserIDENTIFIER          = 42
	CvltParserDELIMITEDIDENTIFIER = 43
	CvltParserQUOTEDIDENTIFIER    = 44
	CvltParserSTRING              = 45
	CvltParserNUMBER              = 46
	CvltParserLONGNUMBER          = 47
	CvltParserWS                  = 48
	CvltParserCOMMENT             = 49
	CvltParserLINE_COMMENT        = 50
)

// CvltParser rules.
const (
	CvltParserRULE_typeSpecifier           = 0
	CvltParserRULE_namedTypeSpecifier      = 1
	CvltParserRULE_modelIdentifier         = 2
	CvltParserRULE_listTypeSpecifier       = 3
	CvltParserRULE_intervalTypeSpecifier   = 4
	CvltParserRULE_tupleTypeSpecifier      = 5
	CvltParserRULE_tupleElementDefinition  = 6
	CvltParserRULE_choiceTypeSpecifier     = 7
	CvltParserRULE_term                    = 8
	CvltParserRULE_ratio                   = 9
	CvltParserRULE_literal                 = 10
	CvltParserRULE_intervalSelector        = 11
	CvltParserRULE_tupleSelector           = 12
	CvltParserRULE_tupleElementSelector    = 13
	CvltParserRULE_instanceSelector        = 14
	CvltParserRULE_instanceElementSelector = 15
	CvltParserRULE_listSelector            = 16
	CvltParserRULE_displayClause           = 17
	CvltParserRULE_codeSelector            = 18
	CvltParserRULE_conceptSelector         = 19
	CvltParserRULE_identifier              = 20
	CvltParserRULE_quantity                = 21
	CvltParserRULE_unit                    = 22
	CvltParserRULE_dateTimePrecision       = 23
	CvltParserRULE_pluralDateTimePrecision = 24
)

// ITypeSpecifierContext is an interface to support dynamic dispatch.
type ITypeSpecifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NamedTypeSpecifier() INamedTypeSpecifierContext
	ListTypeSpecifier() IListTypeSpecifierContext
	IntervalTypeSpecifier() IIntervalTypeSpecifierContext
	TupleTypeSpecifier() ITupleTypeSpecifierContext
	ChoiceTypeSpecifier() IChoiceTypeSpecifierContext

	// IsTypeSpecifierContext differentiates from other interfaces.
	IsTypeSpecifierContext()
}

type TypeSpecifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTypeSpecifierContext() *TypeSpecifierContext {
	var p = new(TypeSpecifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_typeSpecifier
	return p
}

func InitEmptyTypeSpecifierContext(p *TypeSpecifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_typeSpecifier
}

func (*TypeSpecifierContext) IsTypeSpecifierContext() {}

func NewTypeSpecifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TypeSpecifierContext {
	var p = new(TypeSpecifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_typeSpecifier

	return p
}

func (s *TypeSpecifierContext) GetParser() antlr.Parser { return s.parser }

func (s *TypeSpecifierContext) NamedTypeSpecifier() INamedTypeSpecifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INamedTypeSpecifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INamedTypeSpecifierContext)
}

func (s *TypeSpecifierContext) ListTypeSpecifier() IListTypeSpecifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IListTypeSpecifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IListTypeSpecifierContext)
}

func (s *TypeSpecifierContext) IntervalTypeSpecifier() IIntervalTypeSpecifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIntervalTypeSpecifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIntervalTypeSpecifierContext)
}

func (s *TypeSpecifierContext) TupleTypeSpecifier() ITupleTypeSpecifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITupleTypeSpecifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITupleTypeSpecifierContext)
}

func (s *TypeSpecifierContext) ChoiceTypeSpecifier() IChoiceTypeSpecifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IChoiceTypeSpecifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IChoiceTypeSpecifierContext)
}

func (s *TypeSpecifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TypeSpecifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TypeSpecifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitTypeSpecifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) TypeSpecifier() (localctx ITypeSpecifierContext) {
	localctx = NewTypeSpecifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, CvltParserRULE_typeSpecifier)
	p.SetState(55)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CvltParserIDENTIFIER, CvltParserDELIMITEDIDENTIFIER, CvltParserQUOTEDIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(50)
			p.NamedTypeSpecifier()
		}

	case CvltParserT__1:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(51)
			p.ListTypeSpecifier()
		}

	case CvltParserT__4:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(52)
			p.IntervalTypeSpecifier()
		}

	case CvltParserT__5:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(53)
			p.TupleTypeSpecifier()
		}

	case CvltParserT__9:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(54)
			p.ChoiceTypeSpecifier()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// INamedTypeSpecifierContext is an interface to support dynamic dispatch.
type INamedTypeSpecifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIdentifier() []IIdentifierContext
	Identifier(i int) IIdentifierContext

	// IsNamedTypeSpecifierContext differentiates from other interfaces.
	IsNamedTypeSpecifierContext()
}

type NamedTypeSpecifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNamedTypeSpecifierContext() *NamedTypeSpecifierContext {
	var p = new(NamedTypeSpecifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_namedTypeSpecifier
	return p
}

func InitEmptyNamedTypeSpecifierContext(p *NamedTypeSpecifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_namedTypeSpecifier
}

func (*NamedTypeSpecifierContext) IsNamedTypeSpecifierContext() {}

func NewNamedTypeSpecifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NamedTypeSpecifierContext {
	var p = new(NamedTypeSpecifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_namedTypeSpecifier

	return p
}

func (s *NamedTypeSpecifierContext) GetParser() antlr.Parser { return s.parser }

func (s *NamedTypeSpecifierContext) AllIdentifier() []IIdentifierContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdentifierContext); ok {
			len++
		}
	}

	tst := make([]IIdentifierContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdentifierContext); ok {
			tst[i] = t.(IIdentifierContext)
			i++
		}
	}

	return tst
}

func (s *NamedTypeSpecifierContext) Identifier(i int) IIdentifierContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *NamedTypeSpecifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NamedTypeSpecifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NamedTypeSpecifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitNamedTypeSpecifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) NamedTypeSpecifier() (localctx INamedTypeSpecifierContext) {
	localctx = NewNamedTypeSpecifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, CvltParserRULE_namedTypeSpecifier)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(62)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(57)
				p.Identifier()
			}
			{
				p.SetState(58)
				p.Match(CvltParserT__0)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(64)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	{
		p.SetState(65)
		p.Identifier()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IModelIdentifierContext is an interface to support dynamic dispatch.
type IModelIdentifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext

	// IsModelIdentifierContext differentiates from other interfaces.
	IsModelIdentifierContext()
}

type ModelIdentifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyModelIdentifierContext() *ModelIdentifierContext {
	var p = new(ModelIdentifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_modelIdentifier
	return p
}

func InitEmptyModelIdentifierContext(p *ModelIdentifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_modelIdentifier
}

func (*ModelIdentifierContext) IsModelIdentifierContext() {}

func NewModelIdentifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ModelIdentifierContext {
	var p = new(ModelIdentifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_modelIdentifier

	return p
}

func (s *ModelIdentifierContext) GetParser() antlr.Parser { return s.parser }

func (s *ModelIdentifierContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *ModelIdentifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ModelIdentifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ModelIdentifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitModelIdentifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) ModelIdentifier() (localctx IModelIdentifierContext) {
	localctx = NewModelIdentifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, CvltParserRULE_modelIdentifier)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(67)
		p.Identifier()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IListTypeSpecifierContext is an interface to support dynamic dispatch.
type IListTypeSpecifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TypeSpecifier() ITypeSpecifierContext

	// IsListTypeSpecifierContext differentiates from other interfaces.
	IsListTypeSpecifierContext()
}

type ListTypeSpecifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyListTypeSpecifierContext() *ListTypeSpecifierContext {
	var p = new(ListTypeSpecifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_listTypeSpecifier
	return p
}

func InitEmptyListTypeSpecifierContext(p *ListTypeSpecifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_listTypeSpecifier
}

func (*ListTypeSpecifierContext) IsListTypeSpecifierContext() {}

func NewListTypeSpecifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ListTypeSpecifierContext {
	var p = new(ListTypeSpecifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_listTypeSpecifier

	return p
}

func (s *ListTypeSpecifierContext) GetParser() antlr.Parser { return s.parser }

func (s *ListTypeSpecifierContext) TypeSpecifier() ITypeSpecifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITypeSpecifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITypeSpecifierContext)
}

func (s *ListTypeSpecifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ListTypeSpecifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ListTypeSpecifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitListTypeSpecifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) ListTypeSpecifier() (localctx IListTypeSpecifierContext) {
	localctx = NewListTypeSpecifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, CvltParserRULE_listTypeSpecifier)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(69)
		p.Match(CvltParserT__1)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(70)
		p.Match(CvltParserT__2)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(71)
		p.TypeSpecifier()
	}
	{
		p.SetState(72)
		p.Match(CvltParserT__3)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIntervalTypeSpecifierContext is an interface to support dynamic dispatch.
type IIntervalTypeSpecifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TypeSpecifier() ITypeSpecifierContext

	// IsIntervalTypeSpecifierContext differentiates from other interfaces.
	IsIntervalTypeSpecifierContext()
}

type IntervalTypeSpecifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIntervalTypeSpecifierContext() *IntervalTypeSpecifierContext {
	var p = new(IntervalTypeSpecifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_intervalTypeSpecifier
	return p
}

func InitEmptyIntervalTypeSpecifierContext(p *IntervalTypeSpecifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_intervalTypeSpecifier
}

func (*IntervalTypeSpecifierContext) IsIntervalTypeSpecifierContext() {}

func NewIntervalTypeSpecifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IntervalTypeSpecifierContext {
	var p = new(IntervalTypeSpecifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_intervalTypeSpecifier

	return p
}

func (s *IntervalTypeSpecifierContext) GetParser() antlr.Parser { return s.parser }

func (s *IntervalTypeSpecifierContext) TypeSpecifier() ITypeSpecifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITypeSpecifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITypeSpecifierContext)
}

func (s *IntervalTypeSpecifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IntervalTypeSpecifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IntervalTypeSpecifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitIntervalTypeSpecifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) IntervalTypeSpecifier() (localctx IIntervalTypeSpecifierContext) {
	localctx = NewIntervalTypeSpecifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, CvltParserRULE_intervalTypeSpecifier)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(74)
		p.Match(CvltParserT__4)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(75)
		p.Match(CvltParserT__2)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(76)
		p.TypeSpecifier()
	}
	{
		p.SetState(77)
		p.Match(CvltParserT__3)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITupleTypeSpecifierContext is an interface to support dynamic dispatch.
type ITupleTypeSpecifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllTupleElementDefinition() []ITupleElementDefinitionContext
	TupleElementDefinition(i int) ITupleElementDefinitionContext

	// IsTupleTypeSpecifierContext differentiates from other interfaces.
	IsTupleTypeSpecifierContext()
}

type TupleTypeSpecifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTupleTypeSpecifierContext() *TupleTypeSpecifierContext {
	var p = new(TupleTypeSpecifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_tupleTypeSpecifier
	return p
}

func InitEmptyTupleTypeSpecifierContext(p *TupleTypeSpecifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_tupleTypeSpecifier
}

func (*TupleTypeSpecifierContext) IsTupleTypeSpecifierContext() {}

func NewTupleTypeSpecifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TupleTypeSpecifierContext {
	var p = new(TupleTypeSpecifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_tupleTypeSpecifier

	return p
}

func (s *TupleTypeSpecifierContext) GetParser() antlr.Parser { return s.parser }

func (s *TupleTypeSpecifierContext) AllTupleElementDefinition() []ITupleElementDefinitionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITupleElementDefinitionContext); ok {
			len++
		}
	}

	tst := make([]ITupleElementDefinitionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITupleElementDefinitionContext); ok {
			tst[i] = t.(ITupleElementDefinitionContext)
			i++
		}
	}

	return tst
}

func (s *TupleTypeSpecifierContext) TupleElementDefinition(i int) ITupleElementDefinitionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITupleElementDefinitionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITupleElementDefinitionContext)
}

func (s *TupleTypeSpecifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TupleTypeSpecifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TupleTypeSpecifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitTupleTypeSpecifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) TupleTypeSpecifier() (localctx ITupleTypeSpecifierContext) {
	localctx = NewTupleTypeSpecifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, CvltParserRULE_tupleTypeSpecifier)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(79)
		p.Match(CvltParserT__5)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(80)
		p.Match(CvltParserT__6)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(81)
		p.TupleElementDefinition()
	}
	p.SetState(86)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CvltParserT__7 {
		{
			p.SetState(82)
			p.Match(CvltParserT__7)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(83)
			p.TupleElementDefinition()
		}

		p.SetState(88)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(89)
		p.Match(CvltParserT__8)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITupleElementDefinitionContext is an interface to support dynamic dispatch.
type ITupleElementDefinitionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	TypeSpecifier() ITypeSpecifierContext

	// IsTupleElementDefinitionContext differentiates from other interfaces.
	IsTupleElementDefinitionContext()
}

type TupleElementDefinitionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTupleElementDefinitionContext() *TupleElementDefinitionContext {
	var p = new(TupleElementDefinitionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_tupleElementDefinition
	return p
}

func InitEmptyTupleElementDefinitionContext(p *TupleElementDefinitionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_tupleElementDefinition
}

func (*TupleElementDefinitionContext) IsTupleElementDefinitionContext() {}

func NewTupleElementDefinitionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TupleElementDefinitionContext {
	var p = new(TupleElementDefinitionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_tupleElementDefinition

	return p
}

func (s *TupleElementDefinitionContext) GetParser() antlr.Parser { return s.parser }

func (s *TupleElementDefinitionContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *TupleElementDefinitionContext) TypeSpecifier() ITypeSpecifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITypeSpecifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITypeSpecifierContext)
}

func (s *TupleElementDefinitionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TupleElementDefinitionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TupleElementDefinitionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitTupleElementDefinition(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) TupleElementDefinition() (localctx ITupleElementDefinitionContext) {
	localctx = NewTupleElementDefinitionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, CvltParserRULE_tupleElementDefinition)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(91)
		p.Identifier()
	}
	{
		p.SetState(92)
		p.TypeSpecifier()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IChoiceTypeSpecifierContext is an interface to support dynamic dispatch.
type IChoiceTypeSpecifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllTypeSpecifier() []ITypeSpecifierContext
	TypeSpecifier(i int) ITypeSpecifierContext

	// IsChoiceTypeSpecifierContext differentiates from other interfaces.
	IsChoiceTypeSpecifierContext()
}

type ChoiceTypeSpecifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyChoiceTypeSpecifierContext() *ChoiceTypeSpecifierContext {
	var p = new(ChoiceTypeSpecifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_choiceTypeSpecifier
	return p
}

func InitEmptyChoiceTypeSpecifierContext(p *ChoiceTypeSpecifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_choiceTypeSpecifier
}

func (*ChoiceTypeSpecifierContext) IsChoiceTypeSpecifierContext() {}

func NewChoiceTypeSpecifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ChoiceTypeSpecifierContext {
	var p = new(ChoiceTypeSpecifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_choiceTypeSpecifier

	return p
}

func (s *ChoiceTypeSpecifierContext) GetParser() antlr.Parser { return s.parser }

func (s *ChoiceTypeSpecifierContext) AllTypeSpecifier() []ITypeSpecifierContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITypeSpecifierContext); ok {
			len++
		}
	}

	tst := make([]ITypeSpecifierContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITypeSpecifierContext); ok {
			tst[i] = t.(ITypeSpecifierContext)
			i++
		}
	}

	return tst
}

func (s *ChoiceTypeSpecifierContext) TypeSpecifier(i int) ITypeSpecifierContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITypeSpecifierContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITypeSpecifierContext)
}

func (s *ChoiceTypeSpecifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ChoiceTypeSpecifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ChoiceTypeSpecifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitChoiceTypeSpecifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) ChoiceTypeSpecifier() (localctx IChoiceTypeSpecifierContext) {
	localctx = NewChoiceTypeSpecifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, CvltParserRULE_choiceTypeSpecifier)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(94)
		p.Match(CvltParserT__9)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(95)
		p.Match(CvltParserT__2)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(96)
		p.TypeSpecifier()
	}
	p.SetState(101)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CvltParserT__7 {
		{
			p.SetState(97)
			p.Match(CvltParserT__7)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(98)
			p.TypeSpecifier()
		}

		p.SetState(103)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(104)
		p.Match(CvltParserT__3)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITermContext is an interface to support dynamic dispatch.
type ITermContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsTermContext differentiates from other interfaces.
	IsTermContext()
}

type TermContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTermContext() *TermContext {
	var p = new(TermContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_term
	return p
}

func InitEmptyTermContext(p *TermContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_term
}

func (*TermContext) IsTermContext() {}

func NewTermContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TermContext {
	var p = new(TermContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_term

	return p
}

func (s *TermContext) GetParser() antlr.Parser { return s.parser }

func (s *TermContext) CopyAll(ctx *TermContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *TermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TermContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type TupleSelectorTermContext struct {
	TermContext
}

func NewTupleSelectorTermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *TupleSelectorTermContext {
	var p = new(TupleSelectorTermContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *TupleSelectorTermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TupleSelectorTermContext) TupleSelector() ITupleSelectorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITupleSelectorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITupleSelectorContext)
}

func (s *TupleSelectorTermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitTupleSelectorTerm(s)

	default:
		return t.VisitChildren(s)
	}
}

type LiteralTermContext struct {
	TermContext
}

func NewLiteralTermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LiteralTermContext {
	var p = new(LiteralTermContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *LiteralTermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralTermContext) Literal() ILiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteralContext)
}

func (s *LiteralTermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitLiteralTerm(s)

	default:
		return t.VisitChildren(s)
	}
}

type ConceptSelectorTermContext struct {
	TermContext
}

func NewConceptSelectorTermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConceptSelectorTermContext {
	var p = new(ConceptSelectorTermContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *ConceptSelectorTermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConceptSelectorTermContext) ConceptSelector() IConceptSelectorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConceptSelectorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConceptSelectorContext)
}

func (s *ConceptSelectorTermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitConceptSelectorTerm(s)

	default:
		return t.VisitChildren(s)
	}
}

type CodeSelectorTermContext struct {
	TermContext
}

func NewCodeSelectorTermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CodeSelectorTermContext {
	var p = new(CodeSelectorTermContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *CodeSelectorTermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CodeSelectorTermContext) CodeSelector() ICodeSelectorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICodeSelectorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICodeSelectorContext)
}

func (s *CodeSelectorTermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitCodeSelectorTerm(s)

	default:
		return t.VisitChildren(s)
	}
}

type InstanceSelectorTermContext struct {
	TermContext
}

func NewInstanceSelectorTermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InstanceSelectorTermContext {
	var p = new(InstanceSelectorTermContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *InstanceSelectorTermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InstanceSelectorTermContext) InstanceSelector() IInstanceSelectorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInstanceSelectorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInstanceSelectorContext)
}

func (s *InstanceSelectorTermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitInstanceSelectorTerm(s)

	default:
		return t.VisitChildren(s)
	}
}

type IntervalSelectorTermContext struct {
	TermContext
}

func NewIntervalSelectorTermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *IntervalSelectorTermContext {
	var p = new(IntervalSelectorTermContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *IntervalSelectorTermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IntervalSelectorTermContext) IntervalSelector() IIntervalSelectorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIntervalSelectorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIntervalSelectorContext)
}

func (s *IntervalSelectorTermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitIntervalSelectorTerm(s)

	default:
		return t.VisitChildren(s)
	}
}

type ListSelectorTermContext struct {
	TermContext
}

func NewListSelectorTermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ListSelectorTermContext {
	var p = new(ListSelectorTermContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *ListSelectorTermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ListSelectorTermContext) ListSelector() IListSelectorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IListSelectorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IListSelectorContext)
}

func (s *ListSelectorTermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitListSelectorTerm(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) Term() (localctx ITermContext) {
	localctx = NewTermContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, CvltParserRULE_term)
	p.SetState(113)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext()) {
	case 1:
		localctx = NewLiteralTermContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(106)
			p.Literal()
		}

	case 2:
		localctx = NewIntervalSelectorTermContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(107)
			p.IntervalSelector()
		}

	case 3:
		localctx = NewTupleSelectorTermContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(108)
			p.TupleSelector()
		}

	case 4:
		localctx = NewInstanceSelectorTermContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(109)
			p.InstanceSelector()
		}

	case 5:
		localctx = NewListSelectorTermContext(p, localctx)
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(110)
			p.ListSelector()
		}

	case 6:
		localctx = NewCodeSelectorTermContext(p, localctx)
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(111)
			p.CodeSelector()
		}

	case 7:
		localctx = NewConceptSelectorTermContext(p, localctx)
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(112)
			p.ConceptSelector()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRatioContext is an interface to support dynamic dispatch.
type IRatioContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllQuantity() []IQuantityContext
	Quantity(i int) IQuantityContext

	// IsRatioContext differentiates from other interfaces.
	IsRatioContext()
}

type RatioContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRatioContext() *RatioContext {
	var p = new(RatioContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_ratio
	return p
}

func InitEmptyRatioContext(p *RatioContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_ratio
}

func (*RatioContext) IsRatioContext() {}

func NewRatioContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RatioContext {
	var p = new(RatioContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_ratio

	return p
}

func (s *RatioContext) GetParser() antlr.Parser { return s.parser }

func (s *RatioContext) AllQuantity() []IQuantityContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IQuantityContext); ok {
			len++
		}
	}

	tst := make([]IQuantityContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IQuantityContext); ok {
			tst[i] = t.(IQuantityContext)
			i++
		}
	}

	return tst
}

func (s *RatioContext) Quantity(i int) IQuantityContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IQuantityContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IQuantityContext)
}

func (s *RatioContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RatioContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RatioContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitRatio(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) Ratio() (localctx IRatioContext) {
	localctx = NewRatioContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, CvltParserRULE_ratio)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(115)
		p.Quantity()
	}
	{
		p.SetState(116)
		p.Match(CvltParserT__10)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(117)
		p.Quantity()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILiteralContext is an interface to support dynamic dispatch.
type ILiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsLiteralContext differentiates from other interfaces.
	IsLiteralContext()
}

type LiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteralContext() *LiteralContext {
	var p = new(LiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_literal
	return p
}

func InitEmptyLiteralContext(p *LiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_literal
}

func (*LiteralContext) IsLiteralContext() {}

func NewLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LiteralContext {
	var p = new(LiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_literal

	return p
}

func (s *LiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *LiteralContext) CopyAll(ctx *LiteralContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *LiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type TimeLiteralContext struct {
	LiteralContext
}

func NewTimeLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *TimeLiteralContext {
	var p = new(TimeLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *TimeLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TimeLiteralContext) TIME() antlr.TerminalNode {
	return s.GetToken(CvltParserTIME, 0)
}

func (s *TimeLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitTimeLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type NullLiteralContext struct {
	LiteralContext
}

func NewNullLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NullLiteralContext {
	var p = new(NullLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *NullLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NullLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitNullLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type RatioLiteralContext struct {
	LiteralContext
}

func NewRatioLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *RatioLiteralContext {
	var p = new(RatioLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *RatioLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RatioLiteralContext) Ratio() IRatioContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRatioContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRatioContext)
}

func (s *RatioLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitRatioLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type DateTimeLiteralContext struct {
	LiteralContext
}

func NewDateTimeLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DateTimeLiteralContext {
	var p = new(DateTimeLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *DateTimeLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DateTimeLiteralContext) DATETIME() antlr.TerminalNode {
	return s.GetToken(CvltParserDATETIME, 0)
}

func (s *DateTimeLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitDateTimeLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type StringLiteralContext struct {
	LiteralContext
}

func NewStringLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *StringLiteralContext {
	var p = new(StringLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *StringLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StringLiteralContext) STRING() antlr.TerminalNode {
	return s.GetToken(CvltParserSTRING, 0)
}

func (s *StringLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitStringLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type DateLiteralContext struct {
	LiteralContext
}

func NewDateLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DateLiteralContext {
	var p = new(DateLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *DateLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DateLiteralContext) DATE() antlr.TerminalNode {
	return s.GetToken(CvltParserDATE, 0)
}

func (s *DateLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitDateLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type BooleanLiteralContext struct {
	LiteralContext
}

func NewBooleanLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *BooleanLiteralContext {
	var p = new(BooleanLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *BooleanLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BooleanLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitBooleanLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type NumberLiteralContext struct {
	LiteralContext
}

func NewNumberLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NumberLiteralContext {
	var p = new(NumberLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *NumberLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NumberLiteralContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(CvltParserNUMBER, 0)
}

func (s *NumberLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitNumberLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type LongNumberLiteralContext struct {
	LiteralContext
}

func NewLongNumberLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LongNumberLiteralContext {
	var p = new(LongNumberLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *LongNumberLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LongNumberLiteralContext) LONGNUMBER() antlr.TerminalNode {
	return s.GetToken(CvltParserLONGNUMBER, 0)
}

func (s *LongNumberLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitLongNumberLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type QuantityLiteralContext struct {
	LiteralContext
}

func NewQuantityLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *QuantityLiteralContext {
	var p = new(QuantityLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *QuantityLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *QuantityLiteralContext) Quantity() IQuantityContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IQuantityContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IQuantityContext)
}

func (s *QuantityLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitQuantityLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) Literal() (localctx ILiteralContext) {
	localctx = NewLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, CvltParserRULE_literal)
	var _la int

	p.SetState(129)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext()) {
	case 1:
		localctx = NewBooleanLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(119)
			_la = p.GetTokenStream().LA(1)

			if !(_la == CvltParserT__11 || _la == CvltParserT__12) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	case 2:
		localctx = NewNullLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(120)
			p.Match(CvltParserT__13)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewStringLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(121)
			p.Match(CvltParserSTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		localctx = NewNumberLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(122)
			p.Match(CvltParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		localctx = NewLongNumberLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(123)
			p.Match(CvltParserLONGNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		localctx = NewDateTimeLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(124)
			p.Match(CvltParserDATETIME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 7:
		localctx = NewDateLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(125)
			p.Match(CvltParserDATE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 8:
		localctx = NewTimeLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(126)
			p.Match(CvltParserTIME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 9:
		localctx = NewQuantityLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 9)
		{
			p.SetState(127)
			p.Quantity()
		}

	case 10:
		localctx = NewRatioLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 10)
		{
			p.SetState(128)
			p.Ratio()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIntervalSelectorContext is an interface to support dynamic dispatch.
type IIntervalSelectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllLiteral() []ILiteralContext
	Literal(i int) ILiteralContext

	// IsIntervalSelectorContext differentiates from other interfaces.
	IsIntervalSelectorContext()
}

type IntervalSelectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIntervalSelectorContext() *IntervalSelectorContext {
	var p = new(IntervalSelectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_intervalSelector
	return p
}

func InitEmptyIntervalSelectorContext(p *IntervalSelectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_intervalSelector
}

func (*IntervalSelectorContext) IsIntervalSelectorContext() {}

func NewIntervalSelectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IntervalSelectorContext {
	var p = new(IntervalSelectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_intervalSelector

	return p
}

func (s *IntervalSelectorContext) GetParser() antlr.Parser { return s.parser }

func (s *IntervalSelectorContext) AllLiteral() []ILiteralContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILiteralContext); ok {
			len++
		}
	}

	tst := make([]ILiteralContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILiteralContext); ok {
			tst[i] = t.(ILiteralContext)
			i++
		}
	}

	return tst
}

func (s *IntervalSelectorContext) Literal(i int) ILiteralContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteralContext)
}

func (s *IntervalSelectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IntervalSelectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IntervalSelectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitIntervalSelector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) IntervalSelector() (localctx IIntervalSelectorContext) {
	localctx = NewIntervalSelectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, CvltParserRULE_intervalSelector)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(131)
		p.Match(CvltParserT__4)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(132)
		_la = p.GetTokenStream().LA(1)

		if !(_la == CvltParserT__14 || _la == CvltParserT__15) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(133)
		p.Literal()
	}
	{
		p.SetState(134)
		p.Match(CvltParserT__7)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(135)
		p.Literal()
	}
	{
		p.SetState(136)
		_la = p.GetTokenStream().LA(1)

		if !(_la == CvltParserT__16 || _la == CvltParserT__17) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITupleSelectorContext is an interface to support dynamic dispatch.
type ITupleSelectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllTupleElementSelector() []ITupleElementSelectorContext
	TupleElementSelector(i int) ITupleElementSelectorContext

	// IsTupleSelectorContext differentiates from other interfaces.
	IsTupleSelectorContext()
}

type TupleSelectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTupleSelectorContext() *TupleSelectorContext {
	var p = new(TupleSelectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_tupleSelector
	return p
}

func InitEmptyTupleSelectorContext(p *TupleSelectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_tupleSelector
}

func (*TupleSelectorContext) IsTupleSelectorContext() {}

func NewTupleSelectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TupleSelectorContext {
	var p = new(TupleSelectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_tupleSelector

	return p
}

func (s *TupleSelectorContext) GetParser() antlr.Parser { return s.parser }

func (s *TupleSelectorContext) AllTupleElementSelector() []ITupleElementSelectorContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITupleElementSelectorContext); ok {
			len++
		}
	}

	tst := make([]ITupleElementSelectorContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITupleElementSelectorContext); ok {
			tst[i] = t.(ITupleElementSelectorContext)
			i++
		}
	}

	return tst
}

func (s *TupleSelectorContext) TupleElementSelector(i int) ITupleElementSelectorContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITupleElementSelectorContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITupleElementSelectorContext)
}

func (s *TupleSelectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TupleSelectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TupleSelectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitTupleSelector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) TupleSelector() (localctx ITupleSelectorContext) {
	localctx = NewTupleSelectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, CvltParserRULE_tupleSelector)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(139)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CvltParserT__5 {
		{
			p.SetState(138)
			p.Match(CvltParserT__5)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(141)
		p.Match(CvltParserT__6)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(151)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CvltParserT__10:
		{
			p.SetState(142)
			p.Match(CvltParserT__10)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CvltParserIDENTIFIER, CvltParserDELIMITEDIDENTIFIER, CvltParserQUOTEDIDENTIFIER:
		{
			p.SetState(143)
			p.TupleElementSelector()
		}
		p.SetState(148)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CvltParserT__7 {
			{
				p.SetState(144)
				p.Match(CvltParserT__7)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(145)
				p.TupleElementSelector()
			}

			p.SetState(150)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	{
		p.SetState(153)
		p.Match(CvltParserT__8)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITupleElementSelectorContext is an interface to support dynamic dispatch.
type ITupleElementSelectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	Term() ITermContext

	// IsTupleElementSelectorContext differentiates from other interfaces.
	IsTupleElementSelectorContext()
}

type TupleElementSelectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTupleElementSelectorContext() *TupleElementSelectorContext {
	var p = new(TupleElementSelectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_tupleElementSelector
	return p
}

func InitEmptyTupleElementSelectorContext(p *TupleElementSelectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_tupleElementSelector
}

func (*TupleElementSelectorContext) IsTupleElementSelectorContext() {}

func NewTupleElementSelectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TupleElementSelectorContext {
	var p = new(TupleElementSelectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_tupleElementSelector

	return p
}

func (s *TupleElementSelectorContext) GetParser() antlr.Parser { return s.parser }

func (s *TupleElementSelectorContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *TupleElementSelectorContext) Term() ITermContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *TupleElementSelectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TupleElementSelectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TupleElementSelectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitTupleElementSelector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) TupleElementSelector() (localctx ITupleElementSelectorContext) {
	localctx = NewTupleElementSelectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, CvltParserRULE_tupleElementSelector)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(155)
		p.Identifier()
	}
	{
		p.SetState(156)
		p.Match(CvltParserT__10)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(157)
		p.Term()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IInstanceSelectorContext is an interface to support dynamic dispatch.
type IInstanceSelectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	AllInstanceElementSelector() []IInstanceElementSelectorContext
	InstanceElementSelector(i int) IInstanceElementSelectorContext

	// IsInstanceSelectorContext differentiates from other interfaces.
	IsInstanceSelectorContext()
}

type InstanceSelectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInstanceSelectorContext() *InstanceSelectorContext {
	var p = new(InstanceSelectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_instanceSelector
	return p
}

func InitEmptyInstanceSelectorContext(p *InstanceSelectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_instanceSelector
}

func (*InstanceSelectorContext) IsInstanceSelectorContext() {}

func NewInstanceSelectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InstanceSelectorContext {
	var p = new(InstanceSelectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_instanceSelector

	return p
}

func (s *InstanceSelectorContext) GetParser() antlr.Parser { return s.parser }

func (s *InstanceSelectorContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *InstanceSelectorContext) AllInstanceElementSelector() []IInstanceElementSelectorContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IInstanceElementSelectorContext); ok {
			len++
		}
	}

	tst := make([]IInstanceElementSelectorContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IInstanceElementSelectorContext); ok {
			tst[i] = t.(IInstanceElementSelectorContext)
			i++
		}
	}

	return tst
}

func (s *InstanceSelectorContext) InstanceElementSelector(i int) IInstanceElementSelectorContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInstanceElementSelectorContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInstanceElementSelectorContext)
}

func (s *InstanceSelectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InstanceSelectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *InstanceSelectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitInstanceSelector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) InstanceSelector() (localctx IInstanceSelectorContext) {
	localctx = NewInstanceSelectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, CvltParserRULE_instanceSelector)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(159)
		p.Identifier()
	}
	{
		p.SetState(160)
		p.Match(CvltParserT__6)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(170)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CvltParserT__10:
		{
			p.SetState(161)
			p.Match(CvltParserT__10)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CvltParserIDENTIFIER, CvltParserDELIMITEDIDENTIFIER, CvltParserQUOTEDIDENTIFIER:
		{
			p.SetState(162)
			p.InstanceElementSelector()
		}
		p.SetState(167)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CvltParserT__7 {
			{
				p.SetState(163)
				p.Match(CvltParserT__7)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(164)
				p.InstanceElementSelector()
			}

			p.SetState(169)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	{
		p.SetState(172)
		p.Match(CvltParserT__8)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IInstanceElementSelectorContext is an interface to support dynamic dispatch.
type IInstanceElementSelectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	Term() ITermContext

	// IsInstanceElementSelectorContext differentiates from other interfaces.
	IsInstanceElementSelectorContext()
}

type InstanceElementSelectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInstanceElementSelectorContext() *InstanceElementSelectorContext {
	var p = new(InstanceElementSelectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_instanceElementSelector
	return p
}

func InitEmptyInstanceElementSelectorContext(p *InstanceElementSelectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_instanceElementSelector
}

func (*InstanceElementSelectorContext) IsInstanceElementSelectorContext() {}

func NewInstanceElementSelectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InstanceElementSelectorContext {
	var p = new(InstanceElementSelectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_instanceElementSelector

	return p
}

func (s *InstanceElementSelectorContext) GetParser() antlr.Parser { return s.parser }

func (s *InstanceElementSelectorContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *InstanceElementSelectorContext) Term() ITermContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *InstanceElementSelectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InstanceElementSelectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *InstanceElementSelectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitInstanceElementSelector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) InstanceElementSelector() (localctx IInstanceElementSelectorContext) {
	localctx = NewInstanceElementSelectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, CvltParserRULE_instanceElementSelector)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(174)
		p.Identifier()
	}
	{
		p.SetState(175)
		p.Match(CvltParserT__10)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(176)
		p.Term()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IListSelectorContext is an interface to support dynamic dispatch.
type IListSelectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllTerm() []ITermContext
	Term(i int) ITermContext
	TypeSpecifier() ITypeSpecifierContext

	// IsListSelectorContext differentiates from other interfaces.
	IsListSelectorContext()
}

type ListSelectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyListSelectorContext() *ListSelectorContext {
	var p = new(ListSelectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_listSelector
	return p
}

func InitEmptyListSelectorContext(p *ListSelectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_listSelector
}

func (*ListSelectorContext) IsListSelectorContext() {}

func NewListSelectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ListSelectorContext {
	var p = new(ListSelectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_listSelector

	return p
}

func (s *ListSelectorContext) GetParser() antlr.Parser { return s.parser }

func (s *ListSelectorContext) AllTerm() []ITermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITermContext); ok {
			len++
		}
	}

	tst := make([]ITermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITermContext); ok {
			tst[i] = t.(ITermContext)
			i++
		}
	}

	return tst
}

func (s *ListSelectorContext) Term(i int) ITermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *ListSelectorContext) TypeSpecifier() ITypeSpecifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITypeSpecifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITypeSpecifierContext)
}

func (s *ListSelectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ListSelectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ListSelectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitListSelector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) ListSelector() (localctx IListSelectorContext) {
	localctx = NewListSelectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, CvltParserRULE_listSelector)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(185)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CvltParserT__1 {
		{
			p.SetState(178)
			p.Match(CvltParserT__1)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(183)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CvltParserT__2 {
			{
				p.SetState(179)
				p.Match(CvltParserT__2)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(180)
				p.TypeSpecifier()
			}
			{
				p.SetState(181)
				p.Match(CvltParserT__3)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	}
	{
		p.SetState(187)
		p.Match(CvltParserT__6)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(196)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&280925226168548) != 0 {
		{
			p.SetState(188)
			p.Term()
		}
		p.SetState(193)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CvltParserT__7 {
			{
				p.SetState(189)
				p.Match(CvltParserT__7)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(190)
				p.Term()
			}

			p.SetState(195)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	}
	{
		p.SetState(198)
		p.Match(CvltParserT__8)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDisplayClauseContext is an interface to support dynamic dispatch.
type IDisplayClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING() antlr.TerminalNode

	// IsDisplayClauseContext differentiates from other interfaces.
	IsDisplayClauseContext()
}

type DisplayClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDisplayClauseContext() *DisplayClauseContext {
	var p = new(DisplayClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_displayClause
	return p
}

func InitEmptyDisplayClauseContext(p *DisplayClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_displayClause
}

func (*DisplayClauseContext) IsDisplayClauseContext() {}

func NewDisplayClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DisplayClauseContext {
	var p = new(DisplayClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_displayClause

	return p
}

func (s *DisplayClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *DisplayClauseContext) STRING() antlr.TerminalNode {
	return s.GetToken(CvltParserSTRING, 0)
}

func (s *DisplayClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DisplayClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DisplayClauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitDisplayClause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) DisplayClause() (localctx IDisplayClauseContext) {
	localctx = NewDisplayClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, CvltParserRULE_displayClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(200)
		p.Match(CvltParserT__18)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(201)
		p.Match(CvltParserSTRING)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICodeSelectorContext is an interface to support dynamic dispatch.
type ICodeSelectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING() antlr.TerminalNode
	Identifier() IIdentifierContext
	DisplayClause() IDisplayClauseContext

	// IsCodeSelectorContext differentiates from other interfaces.
	IsCodeSelectorContext()
}

type CodeSelectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCodeSelectorContext() *CodeSelectorContext {
	var p = new(CodeSelectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_codeSelector
	return p
}

func InitEmptyCodeSelectorContext(p *CodeSelectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_codeSelector
}

func (*CodeSelectorContext) IsCodeSelectorContext() {}

func NewCodeSelectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CodeSelectorContext {
	var p = new(CodeSelectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_codeSelector

	return p
}

func (s *CodeSelectorContext) GetParser() antlr.Parser { return s.parser }

func (s *CodeSelectorContext) STRING() antlr.TerminalNode {
	return s.GetToken(CvltParserSTRING, 0)
}

func (s *CodeSelectorContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *CodeSelectorContext) DisplayClause() IDisplayClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDisplayClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDisplayClauseContext)
}

func (s *CodeSelectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CodeSelectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CodeSelectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitCodeSelector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) CodeSelector() (localctx ICodeSelectorContext) {
	localctx = NewCodeSelectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 36, CvltParserRULE_codeSelector)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(203)
		p.Match(CvltParserT__19)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(204)
		p.Match(CvltParserSTRING)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(205)
		p.Match(CvltParserT__20)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(206)
		p.Identifier()
	}
	p.SetState(208)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CvltParserT__18 {
		{
			p.SetState(207)
			p.DisplayClause()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IConceptSelectorContext is an interface to support dynamic dispatch.
type IConceptSelectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllCodeSelector() []ICodeSelectorContext
	CodeSelector(i int) ICodeSelectorContext
	DisplayClause() IDisplayClauseContext

	// IsConceptSelectorContext differentiates from other interfaces.
	IsConceptSelectorContext()
}

type ConceptSelectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConceptSelectorContext() *ConceptSelectorContext {
	var p = new(ConceptSelectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_conceptSelector
	return p
}

func InitEmptyConceptSelectorContext(p *ConceptSelectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_conceptSelector
}

func (*ConceptSelectorContext) IsConceptSelectorContext() {}

func NewConceptSelectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ConceptSelectorContext {
	var p = new(ConceptSelectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_conceptSelector

	return p
}

func (s *ConceptSelectorContext) GetParser() antlr.Parser { return s.parser }

func (s *ConceptSelectorContext) AllCodeSelector() []ICodeSelectorContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ICodeSelectorContext); ok {
			len++
		}
	}

	tst := make([]ICodeSelectorContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ICodeSelectorContext); ok {
			tst[i] = t.(ICodeSelectorContext)
			i++
		}
	}

	return tst
}

func (s *ConceptSelectorContext) CodeSelector(i int) ICodeSelectorContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICodeSelectorContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICodeSelectorContext)
}

func (s *ConceptSelectorContext) DisplayClause() IDisplayClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDisplayClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDisplayClauseContext)
}

func (s *ConceptSelectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConceptSelectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ConceptSelectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitConceptSelector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) ConceptSelector() (localctx IConceptSelectorContext) {
	localctx = NewConceptSelectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, CvltParserRULE_conceptSelector)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(210)
		p.Match(CvltParserT__21)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(211)
		p.Match(CvltParserT__6)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(212)
		p.CodeSelector()
	}
	p.SetState(217)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CvltParserT__7 {
		{
			p.SetState(213)
			p.Match(CvltParserT__7)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(214)
			p.CodeSelector()
		}

		p.SetState(219)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(220)
		p.Match(CvltParserT__8)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(222)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CvltParserT__18 {
		{
			p.SetState(221)
			p.DisplayClause()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIdentifierContext is an interface to support dynamic dispatch.
type IIdentifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	DELIMITEDIDENTIFIER() antlr.TerminalNode
	QUOTEDIDENTIFIER() antlr.TerminalNode

	// IsIdentifierContext differentiates from other interfaces.
	IsIdentifierContext()
}

type IdentifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIdentifierContext() *IdentifierContext {
	var p = new(IdentifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_identifier
	return p
}

func InitEmptyIdentifierContext(p *IdentifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_identifier
}

func (*IdentifierContext) IsIdentifierContext() {}

func NewIdentifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IdentifierContext {
	var p = new(IdentifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_identifier

	return p
}

func (s *IdentifierContext) GetParser() antlr.Parser { return s.parser }

func (s *IdentifierContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(CvltParserIDENTIFIER, 0)
}

func (s *IdentifierContext) DELIMITEDIDENTIFIER() antlr.TerminalNode {
	return s.GetToken(CvltParserDELIMITEDIDENTIFIER, 0)
}

func (s *IdentifierContext) QUOTEDIDENTIFIER() antlr.TerminalNode {
	return s.GetToken(CvltParserQUOTEDIDENTIFIER, 0)
}

func (s *IdentifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IdentifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IdentifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitIdentifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) Identifier() (localctx IIdentifierContext) {
	localctx = NewIdentifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 40, CvltParserRULE_identifier)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(224)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&30786325577728) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IQuantityContext is an interface to support dynamic dispatch.
type IQuantityContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NUMBER() antlr.TerminalNode
	Unit() IUnitContext

	// IsQuantityContext differentiates from other interfaces.
	IsQuantityContext()
}

type QuantityContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyQuantityContext() *QuantityContext {
	var p = new(QuantityContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_quantity
	return p
}

func InitEmptyQuantityContext(p *QuantityContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_quantity
}

func (*QuantityContext) IsQuantityContext() {}

func NewQuantityContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *QuantityContext {
	var p = new(QuantityContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_quantity

	return p
}

func (s *QuantityContext) GetParser() antlr.Parser { return s.parser }

func (s *QuantityContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(CvltParserNUMBER, 0)
}

func (s *QuantityContext) Unit() IUnitContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUnitContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUnitContext)
}

func (s *QuantityContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *QuantityContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *QuantityContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitQuantity(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) Quantity() (localctx IQuantityContext) {
	localctx = NewQuantityContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 42, CvltParserRULE_quantity)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(226)
		p.Match(CvltParserNUMBER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(228)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&35734119514112) != 0 {
		{
			p.SetState(227)
			p.Unit()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IUnitContext is an interface to support dynamic dispatch.
type IUnitContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DateTimePrecision() IDateTimePrecisionContext
	PluralDateTimePrecision() IPluralDateTimePrecisionContext
	STRING() antlr.TerminalNode

	// IsUnitContext differentiates from other interfaces.
	IsUnitContext()
}

type UnitContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUnitContext() *UnitContext {
	var p = new(UnitContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_unit
	return p
}

func InitEmptyUnitContext(p *UnitContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_unit
}

func (*UnitContext) IsUnitContext() {}

func NewUnitContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UnitContext {
	var p = new(UnitContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_unit

	return p
}

func (s *UnitContext) GetParser() antlr.Parser { return s.parser }

func (s *UnitContext) DateTimePrecision() IDateTimePrecisionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDateTimePrecisionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDateTimePrecisionContext)
}

func (s *UnitContext) PluralDateTimePrecision() IPluralDateTimePrecisionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPluralDateTimePrecisionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPluralDateTimePrecisionContext)
}

func (s *UnitContext) STRING() antlr.TerminalNode {
	return s.GetToken(CvltParserSTRING, 0)
}

func (s *UnitContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UnitContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UnitContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitUnit(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) Unit() (localctx IUnitContext) {
	localctx = NewUnitContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 44, CvltParserRULE_unit)
	p.SetState(233)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CvltParserT__22, CvltParserT__23, CvltParserT__24, CvltParserT__25, CvltParserT__26, CvltParserT__27, CvltParserT__28, CvltParserT__29:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(230)
			p.DateTimePrecision()
		}

	case CvltParserT__30, CvltParserT__31, CvltParserT__32, CvltParserT__33, CvltParserT__34, CvltParserT__35, CvltParserT__36, CvltParserT__37:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(231)
			p.PluralDateTimePrecision()
		}

	case CvltParserSTRING:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(232)
			p.Match(CvltParserSTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDateTimePrecisionContext is an interface to support dynamic dispatch.
type IDateTimePrecisionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsDateTimePrecisionContext differentiates from other interfaces.
	IsDateTimePrecisionContext()
}

type DateTimePrecisionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDateTimePrecisionContext() *DateTimePrecisionContext {
	var p = new(DateTimePrecisionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_dateTimePrecision
	return p
}

func InitEmptyDateTimePrecisionContext(p *DateTimePrecisionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_dateTimePrecision
}

func (*DateTimePrecisionContext) IsDateTimePrecisionContext() {}

func NewDateTimePrecisionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DateTimePrecisionContext {
	var p = new(DateTimePrecisionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_dateTimePrecision

	return p
}

func (s *DateTimePrecisionContext) GetParser() antlr.Parser { return s.parser }
func (s *DateTimePrecisionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DateTimePrecisionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DateTimePrecisionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitDateTimePrecision(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) DateTimePrecision() (localctx IDateTimePrecisionContext) {
	localctx = NewDateTimePrecisionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 46, CvltParserRULE_dateTimePrecision)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(235)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2139095040) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPluralDateTimePrecisionContext is an interface to support dynamic dispatch.
type IPluralDateTimePrecisionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsPluralDateTimePrecisionContext differentiates from other interfaces.
	IsPluralDateTimePrecisionContext()
}

type PluralDateTimePrecisionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPluralDateTimePrecisionContext() *PluralDateTimePrecisionContext {
	var p = new(PluralDateTimePrecisionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_pluralDateTimePrecision
	return p
}

func InitEmptyPluralDateTimePrecisionContext(p *PluralDateTimePrecisionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CvltParserRULE_pluralDateTimePrecision
}

func (*PluralDateTimePrecisionContext) IsPluralDateTimePrecisionContext() {}

func NewPluralDateTimePrecisionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PluralDateTimePrecisionContext {
	var p = new(PluralDateTimePrecisionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CvltParserRULE_pluralDateTimePrecision

	return p
}

func (s *PluralDateTimePrecisionContext) GetParser() antlr.Parser { return s.parser }
func (s *PluralDateTimePrecisionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PluralDateTimePrecisionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PluralDateTimePrecisionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case CvltVisitor:
		return t.VisitPluralDateTimePrecision(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *CvltParser) PluralDateTimePrecision() (localctx IPluralDateTimePrecisionContext) {
	localctx = NewPluralDateTimePrecisionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 48, CvltParserRULE_pluralDateTimePrecision)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(237)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&547608330240) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
