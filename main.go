package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func main() {
	var expr = &ExprNode{
		Value: "+",
		Left: &ExprNode{
			Value: "1",
		},
		Right: &ExprNode{
			Value: "*",
			Left: &ExprNode{
				Value: "2",
			},
			Right: &ExprNode{
				Value: "+",
				Left: &ExprNode{
					Value: "3",
				},
				Right: &ExprNode{
					Value: "4",
				},
			},
		},
	}
	c := &Compiler{
		nextId: 1,
	}
	s := c.GenLLIR(expr)
	fmt.Println(s)
}

func gen_asm(tokens []string) string {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, `define i32 @main(){`)

	var idx int
	for i, tok := range tokens {
		if i == 0 {
			fmt.Fprintf(&buf, "\t%%t%d = add i32 0, %v\n", idx, tokens[i])
			continue
		}
		switch tok {
		case "+":
			idx++
			fmt.Fprintf(&buf, "\t%%t%d = add i32 %%t%d, %v\n", idx, idx-1, tokens[i+1])
		case "-":
			idx++
			fmt.Fprintf(&buf, "\t%%t%d = sub i32 %%t%d, %v\n", idx, idx-1, tokens[i+1])
		}
	}
	fmt.Fprintf(&buf, "\tret i32 %%t%d\n", idx)
	fmt.Fprintln(&buf, `}`)

	return buf.String()
}

func parse_tokens(code string) (tokens []string) {
	for code != "" {
		if idx := strings.IndexAny(code, "+-"); idx >= 0 {
			if idx > 0 {
				tokens = append(tokens, strings.TrimSpace(code[:idx]))
			}
			tokens = append(tokens, code[idx:][:1])
			code = code[idx+1:]
			continue
		}
		tokens = append(tokens, strings.TrimSpace(code))
		return
	}
	return
}

func compile(code string) {
	tokens := parse_tokens(code)
	output := gen_asm(tokens)

	os.WriteFile("a.out.ll", []byte(output), 0666)
	exec.Command("clang", "-Wno-override-module", "-o", "a.out", "a.out.ll").Run()
}

func run(code string) int {
	compile(code)
	if err := exec.Command("./a.out").Run(); err != nil {
		return err.(*exec.ExitError).ExitCode()
	}
	return 0
}

func TestRun(t *testing.T) {
	for i, tt := range tests {
		if got := run(tt.code); got != tt.value {
			t.Fatalf("%d: expect = %v, got = %v", i, tt.value, got)
		}
	}
}

var tests = []struct {
	code  string
	value int
}{
	{code: "1", value: 1},
	{code: "1+1", value: 2},
	{code: "1 + 3 - 2", value: 2},
	{code: "1 + 2 + 3 + 4", value: 10},
}
