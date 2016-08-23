package templater

import (
	"reflect"
	"errors"
	"strings"
	"log"
)


var transformer_funcs = map[string]interface{} {
            "replace": strings.Replace,
    }


func Call(name string, params ... interface{}) (result []reflect.Value, err error) {
	log.Printf("%v", params)

    f := reflect.ValueOf(transformer_funcs[name])
    if len(params) != f.Type().NumIn() {
		log.Printf("%v", f.Type().NumIn())
        return []reflect.Value{}, errors.New("The number of params is not adapted.")
    }
    in := make([]reflect.Value, len(params))
    for k, param := range params {
        in[k] = reflect.ValueOf(param)
    }

	log.Printf("%v", in)

    result = f.Call(in)
    return result, nil
}

func Transform(s interface{}, cmd string) error {
	c := strings.Split(cmd, "(")
	params := []string{}
	func_name := strings.TrimSpace(c[0])
	if len(c) == 2{
		params = strings.Split(strings.Replace(c[1], ")", "", -1), ",")
	}

	func_params := make([]interface{}, len(params) + 1)
	func_params[0] = s

	for i, v := range params {
    	func_params[i + 1] = v
	}

	log.Println(func_name)
	log.Printf("%v", func_params)

	res, err := Call(func_name, func_params...)

	log.Printf("%v", res)
	log.Printf("%v", err)

	return nil
}