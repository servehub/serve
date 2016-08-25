package templater

import (
	"testing"
	"fmt"
)

func TestParser(t *testing.T) {
	var testData = map[interface{}]interface{}{
		//"true ": "true",
		//"1 ": "1",
		//"true |reverse": "false",
		"var": "var",
		"var1": "var1",
		"var-var": "var-var",
		"var.var": "var.var",
		"var--v |  replace('\\W','_')": "var__v",
		"var--v | replace('\\W','*')": "var**v",
	}



	for input, output := range testData {
		m := Modify{}
		if res, err := m.Exec(fmt.Sprintf("%v",input), nil); err == nil {
			if output == res {
				fmt.Printf("result %v == %v\n", output, res)
			} else {
				fmt.Printf("result %v != %v\n", output, res)
				t.Fail()
			}
		} else {
			fmt.Printf("error %v\n", err)
		}
	}
}
