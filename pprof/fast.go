package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// вам надо написать более быструю оптимальную этой функции
// goos: linux
// goarch: amd64
// pkg: hw3
// cpu: 12th Gen Intel(R) Core(TM) i7-1255U
// BenchmarkSlow-12               2         814577688 ns/op        20155196 B/op     182792 allocs/op
// BenchmarkFast-12              42          26806356 ns/op         1679195 B/op       8490 allocs/op
// PASS
// ok      hw3     5.432s
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	seenBrowsers := map[string]bool{}

	var line string
	var isAndroid bool
	var isMSIE bool
	var user User

	fmt.Fprintln(out, "found users:")

	i := -1
	for scanner.Scan() {
		line = scanner.Text()
		// fmt.Printf("%v %v\n", err, line)
		err := easyjson.Unmarshal([]byte(line), &user)
		if err != nil {
			panic(err)
		}

		isAndroid = false
		isMSIE = false

		for _, browser := range user.Browsers {
			if ok := strings.Contains(browser, "Android"); ok {
				isAndroid = true
				if _, ok := seenBrowsers[browser]; !ok {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers[browser] = true
				}
			} else if ok := strings.Contains(browser, "MSIE"); ok {
				isMSIE = true
				if _, ok := seenBrowsers[browser]; !ok {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers[browser] = true
				}
			}
		}

		i++

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := strings.ReplaceAll(user.Email, "@", " [at] ")
		fmt.Fprintln(out, fmt.Sprintf("[%d] %s <%s>", i, user.Name, email))
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}

//easyjson:json
type User struct {
	Browsers []string `json:"browsers"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
}

func easyjson750e3c9dDecodeHw3Models(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "email":
			out.Email = string(in.String())
		case "name":
			out.Name = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson750e3c9dEncodeHw3Models(out *jwriter.Writer, in User) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"browsers\":"
		out.RawString(prefix[1:])
		if in.Browsers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Browsers {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"email\":"
		out.RawString(prefix)
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v User) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson750e3c9dEncodeHw3Models(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v User) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson750e3c9dEncodeHw3Models(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson750e3c9dDecodeHw3Models(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson750e3c9dDecodeHw3Models(l, v)
}
