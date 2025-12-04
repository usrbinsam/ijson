// Package ijson provides a wrapper around a JSON decoder
// to handle deserializing partial JSON streams.
//
// The ijson package makes streaming JSON output from an
// LLM API useable before the stream is complete.
//
// For example, LLMs in streaming mode will output JSON
// only a few characters at a time.
//
// ```json
// {"
// {"fo
// {"foo"
// {"foo": "b
// ````
// The ijson package works by adding missing paired tokens to the
// stream until a matching one arrives from the stream, similar to
// how a code editor might autocomplete quotes or braces.
//
// This means that in some cases, the JSON stream will be too incomplete
// for the completion technique to produce anything meaningful. In the
// above example, the JSON stream does not have enough partial data until
// the `b` character arrives when the open quote at column 7 and the curly
// bracket at column 0 can be closed by the builder.
package ijson

import (
	"fmt"
	"slices"
	"strings"
)

// A JSONBuilder is used to assemple a JSON string character by character
// using the [JSONBuilder.Write] method. The zero value cannot be used,
// and should be created using [NewJSONBuilder].
// Type T is the type that [JSONBuilder.Value] should return.
type JSONBuilder[T any] struct {
	stringBuilder      strings.Builder
	lifo               []rune
	inQuote            bool
	trailingComma      bool
	trailingCommaIndex int
	escapeNext         bool
	unmarshalFunc      UnmarshalFunc
}

// NewJSONBuilder creates a new JSONBuilder that uses the provided UnmarsahlFunc
// as the decoder for [JSONBuilder.Value].
func NewJSONBuilder[T any](unmarshalFunc UnmarshalFunc) *JSONBuilder[T] {
	return &JSONBuilder[T]{unmarshalFunc: unmarshalFunc}
}

// Value returns the JSON value that was built by the [JSONBuilder.Write] method.
// The error value is whatever the [UnmarshalFunc]'s error value is.
// Expect errors in cases where the JSON stream isn't complete enough yet to
// produce a meaningful value.
func (b JSONBuilder[T]) Value() (T, error) {
	var v T
	data := []byte(b.String())
	err := b.unmarshalFunc(data, &v)
	return v, err
}

// A UnmarshalFunc is any function that implements the same
// call signature as the [encoding/json.Unmarshal] funtion.
type UnmarshalFunc func(data []byte, v any) error

func (b *JSONBuilder[T]) lastAutoClosed() rune {
	return b.lifo[len(b.lifo)-1]
}

// Write process the given characters and appends them to
// the internal JSON string.
func (b *JSONBuilder[T]) Write(in string) {
	b.stringBuilder.Grow(len(in))
	for _, c := range in {
		if b.escapeNext {
			b.stringBuilder.WriteRune(c)
			b.escapeNext = false
			continue
		}
		if c == '\\' {
			if !b.inQuote {
				b.stringBuilder.WriteRune(c)
				panic(fmt.Sprintf("got escape character outside of quote: %s", b.stringBuilder.String()))
			}
			b.escapeNext = true
		}
	S:
		switch c {
		case '{':
			if b.inQuote {
				break S
			}
			b.lifo = append(b.lifo, '}')
			b.trailingComma = false
		case '[':
			if b.inQuote {
				break S
			}
			b.lifo = append(b.lifo, ']')
			b.trailingComma = false
		case '"':
			if b.escapeNext {
				b.escapeNext = false
				break S
			}
			if b.lastAutoClosed() == '"' {
				b.lifo = b.lifo[:len(b.lifo)-1]
				b.inQuote = false
			} else {
				b.lifo = append(b.lifo, '"')
				b.inQuote = true
				b.trailingComma = false
			}
		case '}':
			if b.inQuote {
				break S
			}
			if b.lastAutoClosed() == '}' {
				b.lifo = b.lifo[:len(b.lifo)-1]
			}
			if b.trailingComma {
				panic("received closing brace after a comma")
			}
		case ']':
			if b.inQuote || b.escapeNext {
				break S
			}
			if b.lastAutoClosed() == ']' {
				b.lifo = b.lifo[:len(b.lifo)-1]
			}

			if b.trailingComma {
				panic("received closing bracket after a comma")
			}
		case ',':
			if b.inQuote {
				break S
			}
			b.trailingComma = true
			b.trailingCommaIndex = b.stringBuilder.Len()
		}
		b.stringBuilder.WriteRune(c)
	}
}

// String returns the JSON string that was built by the [JSONBuilder.Write] method.
// Do not expect that the string produced is always valid JSON.
func (b *JSONBuilder[T]) String() string {
	reversed := make([]rune, len(b.lifo))
	_ = copy(reversed, b.lifo)
	slices.Reverse(reversed)
	strbuf := b.stringBuilder.String()
	if b.trailingComma {
		return strings.TrimSpace(strbuf[:b.trailingCommaIndex]) + string(reversed)
	}
	return strings.TrimSpace(strbuf) + string(reversed)
}

// LIFO returns a string of characters used by the builder to
// complete the partial JSON stream received from [JSONBuilder.Write].
// Once these characters arrive on the stream, they're removed from
// the LIFO.
func (b *JSONBuilder[T]) LIFO() string {
	return string(b.lifo)
}
