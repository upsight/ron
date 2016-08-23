package template

import "testing"

func TestTemplateRenderGoSlice(t *testing.T) {
	inp := `{{ $slice := makeSlice "a" 5 "b" }}
{{ range $i, $v := $slice }}{{$i}}    {{$v}}
{{ end }}
abc: >
    123
`

	want := `
0    a
1    5
2    b

abc: >
    123
`

	got, _ := RenderGo("yaml", inp)
	if got != want {
		t.Fatalf(`Template did not render correctly "%s"`, got)
	}
}

func TestTemplateRenderGoDefault(t *testing.T) {
	inp := `{{ get .Env.NOPE "abc" }} def`
	want := `abc def`

	got, err := RenderGo("yaml", inp)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf(`Template did not render correctly "%s"`, got)
	}
}

func TestTemplateRenderBash(t *testing.T) {
	type tableTest struct {
		in  string
		out interface{}
	}

	tableTests := []tableTest{
		// test getting stdout output
		{`{{bash "echo hi" false true false}}`, `hi`},
		// test getting stderr output
		{`{{bash "echo bye 1>&2" false false true}}`, `bye`},
		// test getting exit status code
		{`{{bash "echo hi" true false false}}`, `0`},
		{`{{bash "false" true false false}}`, `1`},
		{`{{bash "exit 11" true false false}}`, `11`},
	}

	for i, test := range tableTests {
		got, err := RenderGo("yaml", test.in)
		if err != nil {
			t.Error(err)
		}
		if got != test.out {
			t.Errorf(`%d input: "%s" want %+v got %+v`, i, test.in, test.out, got)
		}
	}
}

func TestTemplateRenderUnderscore(t *testing.T) {
	type tableTest struct {
		in  string
		out interface{}
	}

	tableTests := []tableTest{
		{`abc-cde`, `abc_cde`},
		{`-abc-cde`, `_abc_cde`},
		{`abccde`, `abccde`},
	}

	for i, test := range tableTests {
		got := underscore(test.in)
		if got != test.out {
			t.Errorf(`%d input: "%s" want %+v got %+v`, i, test.in, test.out, got)
		}
	}
}
