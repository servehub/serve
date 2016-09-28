package templater

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/InnovaCo/serve/utils/gabs"
	"github.com/InnovaCo/serve/utils/templater/lexer"
	"github.com/InnovaCo/serve/utils/templater/token"
)

var ModifyFuncs = map[string]interface{}{
	"replace": replace,
	"same":    same,
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

func (m Modify) SetFunc(name string, function interface{}) error {
	if _, ok := ModifyFuncs[name]; ok {
		return fmt.Errorf("function %s exist", name)
	}
	ModifyFuncs[name] = function
	return nil
}

func (m Modify) Call(name string, params ...interface{}) (reflect.Value, error) {
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

func (_ Modify) convert(val string) interface{} {
	if val == "" {
		return val
	} else if i, err := strconv.Atoi(val); err == nil {
		return i
	} else if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f
	} else if b, err := strconv.ParseBool(val); err == nil {
		return b
	}
	b := []byte(val)
	if (b[0] == []byte("\"")[0] && b[len(b)-1] == []byte("\"")[0]) ||
		(b[0] == []byte("'")[0] && b[len(b)-1] == []byte("'")[0]) {
		return string(b[1 : len(val)-1])
	}
	return fmt.Sprintf("%v", val)
}

func (_ Modify) clearFunc(s []byte) []string {
	return strings.Split(strings.TrimSpace(string([]byte(strings.TrimSpace(string(s)))[1:])), "(")
}

func (_ Modify) clearArg(s []byte) string {
	a := []byte(strings.TrimSpace(string(s)))
	if a[0] == []byte(",")[0] {
		return strings.TrimSpace(string(a[1:]))
	}
	return string(a)
}

func (m Modify) parseFunc(s []byte) (string, []interface{}, error) {
	f := m.clearFunc(s)
	funcName := f[0]
	funcArgs := []interface{}{nil}
	if len(f) == 1 {
		return funcName, funcArgs, nil
	}
	fl := lexer.NewLexer([]byte(f[1]))
	for ftok := fl.Scan(); ftok.Type == token.TokMap.Type("arg"); ftok = fl.Scan() {
		funcArgs = append(funcArgs, m.convert(m.clearArg(ftok.Lit)))
	}
	return funcName, funcArgs, nil
}

func (m Modify) resolve(v string) (string, error) {
	//fmt.Printf("--> resolve: %v\n", v)
	if m.context == nil {
		//fmt.Printf("<-- resolve: %v\n", v)
		return v, nil
	}
	if value := m.context.Path(v).Data(); value != nil {
		//fmt.Printf("find: %v\n", value)
		v = fmt.Sprintf("%v", value)
	} else {
		//fmt.Println(this.context.String())
		//fmt.Printf("<-- resolve: %v\n", v)
		return v, nil
	}
	if s, err := Template(v, m.context); err != nil {
		//fmt.Println("<-- fuck", v)
		return v, nil
	} else {
		//fmt.Println("<--", v)
		return s, nil
	}
}

func (m Modify) Exec(s string) (interface{}, error) {
	l := lexer.NewLexer([]byte(s))
	var res interface{}
	res = nil

	for tok := l.Scan(); (tok.Type == token.TokMap.Type("var")) ||
		(tok.Type == token.TokMap.Type("func")); tok = l.Scan() {
		switch {
		case tok.Type == token.TokMap.Type("var"):
			//fmt.Printf("var token: %v\n", string(tok.Lit))
			if val, err := m.resolve(string(tok.Lit)); err != nil {
				return nil, err
			} else {
				res = m.convert(val)
			}
		case tok.Type == token.TokMap.Type("func"):
			//fmt.Printf("func token: %v\n", string(tok.Lit))
			if funcName, funcArgs, err := m.parseFunc(tok.Lit); err == nil {
				funcArgs[0] = res
				if fv, err := m.Call(funcName, funcArgs...); err != nil {
					return nil, fmt.Errorf("execution error %s: %v", funcName, err)
				} else {
					res = m.convert(fv.String())
				}
			} else {
				return nil, fmt.Errorf("error parse %s: %v", tok.Lit, err)
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
		return Modify{context}.Exec(fmt.Sprintf("%v", s))
	}
	return s, nil
}
