package templater

import (
	"testing"
	"github.com/InnovaCo/serve/utils/templater/lexer"
	"github.com/InnovaCo/serve/utils/templater/token"
	"fmt"
)

func TestParser(t *testing.T) {
	var testData = map[string]bool{
		//"var1": true,
		//"var1 | var2": true,
		//"func(var2)": true,
		"func()": true,
		"var1 | f(var2,var3)": true,
		//"\"var1\" | f(\"var2\",\"var3\")": true,
		"\"v.var1\" | f(\"v.var2\",\"var3\")": true,
	}


	for input, _ := range testData {
		fmt.Println(input)

		l := lexer.NewLexer([]byte(input))
		for tok := l.Scan(); (tok.Type == token.TokMap.Type("var")) ||
			                 (tok.Type == token.TokMap.Type("func")); tok = l.Scan() {
			switch {
			case  tok.Type == token.TokMap.Type("func"):
				fmt.Println(string(tok.Lit))
				fl := lexer.NewLexer([]byte(tok.Lit))
				for ftok := fl.Scan(); (ftok.Type == token.TokMap.Type("var")); ftok = fl.Scan() {
					fmt.Println(string(ftok.Lit))
				}
			default:
				//fmt.Println(tok.Type)
				fmt.Println(string(tok.Lit))
			}
		}
	}
}
