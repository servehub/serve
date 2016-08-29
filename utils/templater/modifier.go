package templater

import (
	"reflect"
	"errors"
	"strings"
	"log"
	"strconv"
	"fmt"
	"regexp"

	"github.com/InnovaCo/serve/utils/templater/lexer"
	"github.com/InnovaCo/serve/utils/templater/token"
	"github.com/InnovaCo/serve/utils/gabs"
)

var ModifyFuncs = map[string]interface{}{
	"replace": replace,
	"same": same,
	"reverse": reverse,
}

func replace(old, r, new string) string {
	return regexp.MustCompile(r).ReplaceAllString(old, new)
}

func same(s string) string {
	return s
}

func reverse(s bool) bool {
	return !s
}

type Modify struct {
	context *gabs.Container
}

func (this Modify) SetFunc(name string, function interface{}) error {
	if _, ok := ModifyFuncs[name]; ok {
		return fmt.Errorf("function %s exist", name)
	}
	ModifyFuncs[name] = function
	return nil
}

func (this Modify) Call(name string, params ... interface{}) (reflect.Value, error) {
	log.Printf("modify call: func=%s args=%v\n", name, params)

	if _, ok := ModifyFuncs[name]; !ok {
		return reflect.Value{}, fmt.Errorf("function %v not register", name)
	}

    f := reflect.ValueOf(ModifyFuncs[name])
    if len(params) != f.Type().NumIn() {
		return reflect.Value{}, errors.New("The number of params is not adapted.")
    }
    in := make([]reflect.Value, len(params))
    for k, param := range params {
        in[k] = reflect.ValueOf(param)
    }
	return f.Call(in)[0], nil
}

func (this Modify) convert(val string) interface{} {
	if i, err := strconv.Atoi(val); err == nil {
		return i
	} else if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f
	} else  if b, err := strconv.ParseBool(val); err == nil {
		return b
	}
	b := []byte(val)
	if (b[0] == []byte("\"")[0] && b[len(b) - 1] == []byte("\"")[0]) ||
		(b[0] == []byte("'")[0] && b[len(b) - 1] == []byte("'")[0])	{
		return string(b[1:len(val)-1])
	}
	return fmt.Sprintf("%v", val)
}

func (this Modify) clearFunc(s []byte) []string {
	return strings.Split(strings.TrimSpace(string([]byte(strings.TrimSpace(string(s)))[1:])), "(")
}

func (this Modify) clearArg(s []byte) string {
	a := []byte(strings.TrimSpace(string(s)))
	if a[0] == []byte(",")[0] {
		return strings.TrimSpace(string(a[1:]))
	}
	return string(a)
}

func (this Modify) parseFunc(s []byte) (string, []interface{}, error) {
	f := this.clearFunc(s)
	funcName := f[0]
	funcArgs := []interface{}{nil}
	if len(f) == 1 {
		return funcName, funcArgs, nil
	}
	fl := lexer.NewLexer([]byte(f[1]))
	for ftok := fl.Scan(); ftok.Type == token.TokMap.Type("arg"); ftok = fl.Scan() {
		funcArgs = append(funcArgs, this.convert(this.clearArg(ftok.Lit)))
	}
	return funcName, funcArgs, nil
}

func (this Modify) resolve(v string) (string, error) {
	//fmt.Printf("--> resolve: %v\n", v)
	if this.context == nil {
		//fmt.Printf("<-- resolve: %v\n", v)
		return v, nil
	}
	if value := this.context.Path(v).Data(); value != nil {
		//fmt.Printf("find: %v\n", value)
		v = fmt.Sprintf("%v", value)
	} else {
		//fmt.Println(this.context.String())
		//fmt.Printf("<-- resolve: %v\n", v)
		return v, nil
	}
	if s, err := Template(v, this.context); err != nil {
		//fmt.Println("<-- fuck", v)
		return v, nil
	} else {
		//fmt.Println("<--", v)
		return s, nil
	}
}

func (this Modify) Exec(s string) (interface{}, error) {
	l := lexer.NewLexer([]byte(s))
	var res interface{}
	res = nil

	for tok := l.Scan(); (tok.Type == token.TokMap.Type("var")) ||
						 (tok.Type == token.TokMap.Type("func")); tok = l.Scan() {
		switch {
			case tok.Type == token.TokMap.Type("var"):
				//fmt.Printf("var token: %v\n", string(tok.Lit))
				if val, err := this.resolve(string(tok.Lit)); err != nil {
					return nil, err
				} else {
					res = this.convert(val)
				}
			case tok.Type == token.TokMap.Type("func"):
				//fmt.Printf("func token: %v\n", string(tok.Lit))
				if funcName, funcArgs, err := this.parseFunc(tok.Lit); err == nil {
					funcArgs[0] = res
					if fv, err := this.Call(funcName, funcArgs...); err != nil {
						return nil, fmt.Errorf("execution error %s: %v", funcName, err)
					} else {
						res = this.convert(fv.String())
					}
				} else {
					return nil, fmt.Errorf("error parse %s: %v", tok.Lit, err )
				}
			default:
				return nil, fmt.Errorf("unknown token %v\n", string(tok.Lit))
		}
	}
	return res, nil
}

func ModifyExec(s interface{}, context *gabs.Container) (interface{}, error) {
	switch s.(type) {
		case string:
			if strings.Contains(s.(string), "{{") && strings.Contains(s.(string), "}}") {
				return nil, fmt.Errorf("find symbols '{{' and '}}' in %v", s)
			}
			return Modify{context}.Exec(fmt.Sprintf("%v",s))
	}
	return s, nil
}