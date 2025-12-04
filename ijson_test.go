package ijson_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/usrbinsam/ijson"
)

type testStruct struct {
	Foo    string   `json:"foo"`
	Things []string `json:"things"`
}

func Test_JSONBuilder(t *testing.T) {
	testCases := []struct {
		input     string
		output    string
		shouldErr bool
		model     testStruct
	}{
		{
			input:     "{",
			output:    "{}",
			shouldErr: false,
			model:     testStruct{},
		},
		{
			input:     `{}`,
			output:    `{}`,
			shouldErr: false,
			model:     testStruct{},
		},
		{
			input:     `{"foo`,
			output:    `{"foo"}`,
			shouldErr: true,
			model:     testStruct{},
		},
		{
			input:     `{"foo": "ba`,
			output:    `{"foo": "ba"}`,
			shouldErr: false,
			model:     testStruct{Foo: "ba"},
		},
		{
			input:     `{"foo": "bar", "things": ["a"   ,    `,
			output:    `{"foo": "bar", "things": ["a"]}`,
			shouldErr: false,
			model:     testStruct{Foo: "bar", Things: []string{"a"}},
		},
		{
			input:     `{"foo": "bar\"`,
			output:    `{"foo": "bar\""}`,
			shouldErr: false,
			model:     testStruct{Foo: "bar\""},
		},
		{
			input:     `{"foo": "bar\\"`,
			output:    `{"foo": "bar\\"}`,
			shouldErr: false,
			model:     testStruct{Foo: "bar\\"},
		},
		{
			input:     `{"foo":"console.log(\"Mining block \" + (i + 1));`,
			output:    `{"foo":"console.log(\"Mining block \" + (i + 1));"}`,
			shouldErr: false,
			model:     testStruct{Foo: `console.log("Mining block " + (i + 1));`},
		},
		{
			input:     `{"foo":"abc","things":["aaa","bbb","ccc"]}`,
			output:    `{"foo":"abc","things":["aaa","bbb","ccc"]}`,
			shouldErr: false,
			model:     testStruct{Foo: "abc", Things: []string{"aaa", "bbb", "ccc"}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			builder := ijson.NewJSONBuilder[testStruct](json.Unmarshal)
			builder.Write(tc.input)

			if builder.String() != tc.output {
				t.Errorf("Expected %s, got %s", tc.output, builder.String())
			}

			v, err := builder.Value()
			didErr := err != nil

			if tc.shouldErr != didErr {
				t.Errorf("expected err = %v, got err = %v", tc.shouldErr, err.Error())
			}

			if !reflect.DeepEqual(tc.model, v) {
				t.Errorf("expected %+v, got %+v - lifo = %v", tc.model, v, builder.LIFO())
			}
		})
	}
}

//go:embed partial_test.txt
var s string

func Test_JSONBuilder_01(t *testing.T) {
	builder := ijson.JSONBuilder[testStruct]{}
	builder.Write(s)
}
