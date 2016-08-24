package templater

import (
	"testing"
	"fmt"
	"log"
	"github.com/fatih/color"

	"github.com/InnovaCo/serve/utils/gabs"
)

func TestParser(t *testing.T) {
	var testData = map[string]string{
		"var2": "var2",
		"v.var1": "v.var1",
		"var2 | same": "var2",
		"var1 | replace(-,_,-1)": "a_b1",
		"var1 | same": "a-b1",
		"var1 | same | replace(\"-\",\"_\",1)": "a_b1",
		"var1 | p(\"_\",1)": "a_b1",
	}

	tree, _ := gabs.ParseJSON([]byte("{\"var1\": \"a-b1\", \"r\": 1}"))
	log.Printf(tree.String())

	for input, output := range testData {
		if result, err := modify_exec(input, tree); err != nil {
			fmt.Printf("error: %v\n", err)
		} else {
			if output != result {
				color.Red("%v != %v: Error\n", output, result)
				t.Fail()
			} else {
				color.Green("%v: OK\n", result)
			}
		}
	}
}
