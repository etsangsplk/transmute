package parser

import (
	"testing"
	"github.com/hscells/transmute/lexer"
)

var (
	medlineQueryString = `1. exp Sleep Apnea Syndromes/
2. (sleep$ adj3 (apnea$ or apnoea$)).mp.
3. (hypopnoea$ or hypopnoea$).mp.
4. OSA.mp.
5. SHS.mp.
6. OSAHS.mp.
7. or/1-6`
)

func TestBooleanQuery_Terms(t *testing.T) {
	ast, err := lexer.Lex(medlineQueryString)
	if err != nil {
		t.Fatal(err)
	}
	queryRep := NewMedlineParser().Parse(ast)
	if err != nil {
		t.Fatal(err)
	}

	expected := 9
	got := len(queryRep.Terms())
	if expected != got {
		t.Fatalf("Expected %v terms, got %v", expected, got)
	}
}

func TestBooleanQuery_Fields(t *testing.T) {
	ast, err := lexer.Lex(medlineQueryString)
	if err != nil {
		t.Fatal(err)
	}
	queryRep := NewMedlineParser().Parse(ast)
	if err != nil {
		t.Fatal(err)
	}

	expected := 9
	got := len(queryRep.Fields())
	if expected != got {
		t.Fatalf("Expected %v fields, got %v", expected, got)
	}
}



func TestBooleanQuery_FieldCount(t *testing.T) {
	ast, err := lexer.Lex(medlineQueryString)
	if err != nil {
		t.Fatal(err)
	}
	queryRep := NewMedlineParser().Parse(ast)
	if err != nil {
		t.Fatal(err)
	}

	expected := 9
	got := queryRep.FieldCount()["mesh_headings"]
	if expected != got {
		t.Fatalf("Expected %v fields, got %v", expected, got)
	}
}