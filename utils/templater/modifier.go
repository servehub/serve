package templater

import (
	"reflect"
	"errors"
	"strings"
	"log"
	"strconv"
	"fmt"

	"github.com/InnovaCo/serve/utils/gabs"

	"github.com/InnovaCo/serve/utils/templater/lexer"
	"github.com/InnovaCo/serve/utils/templater/token"
)


type Modify struct {
}

var modify_funcs = map[string]interface{}{
	"replace": strings.Replace,
	"same": same,
}


func same(s string) string {
	return s
}

func (p Modify) Call(name string, params ... interface{}) (reflect.Value, error) {
	log.Printf("--> call: %s %v\n", name, params)

	if _, ok := modify_funcs[name]; !ok {
		return reflect.Value{}, fmt.Errorf("function %v not register", name)
	}

    f := reflect.ValueOf(modify_funcs[name])
    if len(params) != f.Type().NumIn() {
		return reflect.Value{}, errors.New("The number of params is not adapted.")
    }
    in := make([]reflect.Value, len(params))
    for k, param := range params {
        in[k] = reflect.ValueOf(param)
    }
	return f.Call(in)[0], nil
}

func (p Modify) convert(val string) interface{} {
	if i, err := strconv.Atoi(val); err == nil {
		//log.Printf("%v is int", val)
		return i
	} else if f, err := strconv.ParseFloat(val, 64); err == nil {
		//log.Printf("%v is float", f)
		return f
	} else  if b, err := strconv.ParseBool(val); err == nil {
		//log.Printf("%v is bool", val)
		return b
	}
	//log.Printf("%v is str", val)
	b := []byte(val)
	if b[0] == []byte("\"")[0] && b[len(b) - 1] == []byte("\"")[0] {
		return string(b[1:len(val)-1])
	}
	return fmt.Sprintf("%v", val)
}


func (p Modify) ParseFunc(s []byte, context *gabs.Container) (string, []interface{}, error) {
	f := strings.Split(string(s), "(")
	func_name := f[0]
	fl := lexer.NewLexer([]byte(f[1])[:len(f[1])-1])
	func_args := []interface{}{nil}

	//log.Printf("func %s, args %v\n", func_name, string([]byte(f[1])[:len(f[1])-1]))

	for ftok := fl.Scan(); ftok.Type == token.TokMap.Type("var"); ftok = fl.Scan() {
		//log.Printf("%v\n", string(ftok.Lit))

		if (context != nil) && context.ExistsP(string(ftok.Lit)) {
			//log.Printf("substitution")
			func_args = append(func_args, p.convert(context.Path(string(ftok.Lit)).String()))
		} else {
			func_args = append(func_args, ftok.Lit)
		}
		val := string(func_args[len(func_args)-1].([]byte))
		//fmt.Println(val)

		func_args[len(func_args)-1] = p.convert(val)

		//log.Printf("%v\n", func_args)
	}
	return func_name, func_args, nil
}

func (p Modify) Exec(s string, context *gabs.Container) (interface{}, error) {
	l := lexer.NewLexer([]byte(s))
	v := []interface{}{nil}
	init := false

	for tok := l.Scan(); (tok.Type == token.TokMap.Type("var")) ||
						 (tok.Type == token.TokMap.Type("func")); tok = l.Scan() {
		//log.Printf("--> %v\n", v)

		switch {
		case  tok.Type == token.TokMap.Type("func"):
			//log.Printf("func %v\n", string(tok.Lit))

			if func_name, func_args, err := p.ParseFunc(tok.Lit, context); err == nil {
				func_args[0] = v[0]
				//log.Printf("%v\n", func_args)

				if fv, err := p.Call(func_name, func_args...); err != nil {
					return nil, fmt.Errorf("execution error %s: %v", func_name, err)
				} else {
					//log.Printf("<-- %s %v\n", func_name, fv)
					v[0] = p.convert(fv.String())
				}
			} else {
				return nil, fmt.Errorf("error parse %s: %v", tok.Lit, err )
			}

		case  tok.Type == token.TokMap.Type("var"):
			if init {
				//log.Printf("func %v\n", string(tok.Lit))

				func_name := string(tok.Lit)
				func_args := []interface{}{v[0]}
				if fv, err := p.Call(func_name, func_args...); err != nil {
					return nil, fmt.Errorf("error executin %s: %v", func_name, err)
				} else {
					//log.Printf("<-- %s %v\n", func_name, fv)
					v[0] = p.convert(fv.String())
				}
			} else {
				//log.Printf("var %v\n", string(tok.Lit))

				if (context != nil) && context.ExistsP(string(tok.Lit)) {
					//log.Printf("substitution")
					v = []interface{}{p.convert(context.Path(string(tok.Lit)).String())}
				} else {
					v = []interface{}{string(tok.Lit)}
				}
				init = true
			}
		default:
			fmt.Println(string(tok.Lit))
		}
	}
	return v[0], nil
}

func modify_exec(s string, context *gabs.Container) (interface{}, error) {
	m := Modify{}
	if result, err := m.Exec(s, context); err != nil {
		return nil, err
	} else {
		return result, err
	}
}
