package parser

import (
	"encoding/json"
	"fmt"
)

func parse(jobs) {
output, err := json.MarshalIndent(jobs, "", "  ")
if err != nil {
    panic(err)
}
fmt.Println(string(output))
}

