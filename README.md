# ijson

ijson is a wrapper around the `json` package that handles incomplete
JSON streams, typically from Large Language Model REST APIs.

In JSON-mode (or Structured Output mode), JSON arrives in only
a few tokens at a time, which is unusable by the standard library
parser until fully assembled. This defeats the purpose of streaming
at all.

`ijson` provides a builder interface that attempts to "complete" the JSON
stream by supplementing unclosed quotes and braces until they arrive on
the stream.

## Usage

1. Declare a struct that matches the JSON to be received.
2. Create a new builder using `ijson.NewJSONBuilder` with your `UnmarshalFunc`
   of choice. The examples use the standard library.
3. As JSON characters arrive from the stream, add them using `JSONBuilder.Write`.
4. Call `JSONBuilder.Value` to see if the builder has enough JSON to produce
   a partially filled instance of the struct.

## Example

The following contrived example assumes existing setup to receive JSON from
an external source in a loop until that source indicates its finished sending.

After each receive, the example attempts to print the struct representation
of the JSON stream. Decoder errors are expected during the early parts of the
stream, when the JSON document doesn't have at least one partial key-value.

```go
import "encoding/json"

type MyType struct {
  Foo string `json:"foo"`
  Bar string `json:"bar"`
}

builder := ijson.NewJSONBuilder[MyType](json.Unmarshal)

llmChan := make(chan string)
go receiveFromLLM(llmChan)

for {
  v, ok := <-llmChan

  if !ok {
    break
  }
  builder.Write(v)

  myType, err := builder.Value()
  if err != nil {
    continue
  }
  log.Printf("LLM: %+v", myType)
}

log.Printf("Final Output: %+v", myType)
```
