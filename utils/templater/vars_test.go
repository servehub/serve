package templater

import (
	"testing"
	"fmt"
	"github.com/InnovaCo/serve/utils/gabs"
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
		"var--v | replace('\\W',  '*')": "var**v",
		"vars2 | replace('\\W',  '*')": "value*unknown",
	}

	json := `{"vars": "value-unknown", "vars1": "{{ vars }}", "vars2": "{{ vars1 }}"}`

	tree, _ := gabs.ParseJSON([]byte(json))

	for input, output := range testData {
		m := Modify{tree}
		if res, err := m.Exec(fmt.Sprintf("%v",input)); err == nil {
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
