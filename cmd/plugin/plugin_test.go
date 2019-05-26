package plugin

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gqlc/compiler"
	"github.com/gqlc/compiler/plugin"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/parser"
	"github.com/gqlc/graphql/token"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func helperCommand(t *testing.T, s ...string) (cmd *exec.Cmd) {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	cmd = exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

const (
	testGql = `scalar Test`
	outDoc  = `Doc received: test, Opts: hello="world!"`
)

var (
	testDoc *ast.Document
)

func TestMain(m *testing.M) {
	var err error
	testDoc, err = parser.ParseDoc(token.NewDocSet(), "test", strings.NewReader(testGql), 0)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

type testCtx struct {
	io.Writer
}

func (ctx testCtx) Open(filename string) (io.WriteCloser, error) { return ctx, nil }

func (ctx testCtx) Close() error { return nil }

func TestGenerator_Generate(t *testing.T) {
	// Get helper cmd
	cmd := helperCommand(t, "generate")

	// Create generate and run generate
	var b bytes.Buffer
	g := &Generator{
		Name: "test",
		Cmd:  cmd,
	}
	ctx := compiler.WithContext(context.Background(), &testCtx{Writer: &b})
	err := g.Generate(ctx, testDoc, `{hello: "world!"}`)
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.EqualFold(b.Bytes(), []byte(outDoc)) {
		t.Fail()
		return
	}
}

func TestUnknownPlugin(t *testing.T) {
	g := &Generator{Name: "nonexistent", Prefix: "gqlc-gen-"}

	err1 := g.Generate(nil, &ast.Document{Name: "Test"}, "")
	if err1 == nil {
		t.Fail()
		return
	}

	err2 := g.Generate(nil, &ast.Document{Name: "Test"}, "")
	if err2 == nil {
		t.Fail()
		return
	}

	ce1, ce2 := err1.(compiler.GeneratorError), err2.(compiler.GeneratorError)
	if ce1.Msg != ce2.Msg {
		t.Fail()
	}
}

func TestMalformedResponse(t *testing.T) {
	// Get helper cmd
	cmd := helperCommand(t, "malformed")

	// Create generate and run generate
	var b bytes.Buffer
	g := &Generator{
		Name: "test",
		Cmd:  cmd,
	}
	ctx := compiler.WithContext(context.Background(), &testCtx{Writer: &b})
	err := g.Generate(ctx, testDoc, "")
	if err == nil {
		t.Error(err)
		return
	}

	cerr, ok := err.(compiler.GeneratorError)
	if !ok {
		t.Fatal("unexpected err type")
		return
	}

	if cerr.Msg != io.ErrUnexpectedEOF.Error() {
		t.Fail()
	}
}

func TestResponseError(t *testing.T) {
	// Get helper cmd
	cmd := helperCommand(t, "error")

	// Create generate and run generate
	var b bytes.Buffer
	g := &Generator{
		Name: "test",
		Cmd:  cmd,
	}
	ctx := compiler.WithContext(context.Background(), &testCtx{Writer: &b})
	err := g.Generate(ctx, testDoc, "")
	if err == nil {
		t.Error(err)
		return
	}

	cerr, ok := err.(compiler.GeneratorError)
	if !ok {
		t.Fatal("unexpected err type")
		return
	}

	if cerr.Msg != "testing error response" {
		t.Fail()
	}
}

// TestHelperProcess isn't a real test. It's used as a helper process
// for TestParameterRun.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "generate":
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}

		var req plugin.Request
		err = proto.Unmarshal(b, &req)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}

		if len(req.FileToGenerate) != 1 {
			fmt.Fprintln(os.Stdout, "expected one file")
			os.Exit(0)
		}

		resp := &plugin.Response{
			File: []*plugin.Response_File{
				{
					Name:    "test.txt",
					Content: outDoc,
				},
			},
		}
		b, err = proto.Marshal(resp)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}

		_, err = os.Stdout.Write(b)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}
	case "malformed":
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}

		var req plugin.Request
		err = proto.Unmarshal(b, &req)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}

		fmt.Println()
		os.Exit(0)
	case "error":
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}

		var req plugin.Request
		err = proto.Unmarshal(b, &req)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}

		if len(req.FileToGenerate) != 1 {
			fmt.Fprintln(os.Stdout, "expected one file")
			os.Exit(0)
		}

		resp := &plugin.Response{
			Error: "testing error response",
		}
		b, err = proto.Marshal(resp)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}

		_, err = os.Stdout.Write(b)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}
	}
}
