package plugin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/gqlc/gqlc/gen"
	"github.com/gqlc/gqlc/plugin/pb"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/parser"
	"github.com/gqlc/graphql/token"
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

func TestGenerator_Generate(t *testing.T) {
	// Get helper cmd
	cmd := helperCommand(t, "generate")

	// Create generate and run generate
	var b bytes.Buffer
	g := &Generator{
		Name: "test",
		Cmd:  cmd,
	}
	ctx := gen.WithContext(context.Background(), gen.TestCtx{Writer: &b})
	err := g.Generate(ctx, testDoc, map[string]interface{}{"hello": "world!"})
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

	err1 := g.Generate(nil, &ast.Document{Name: "Test"}, nil)
	if err1 == nil {
		t.Fail()
		return
	}

	err2 := g.Generate(nil, &ast.Document{Name: "Test"}, nil)
	if err2 == nil {
		t.Fail()
		return
	}

	ce1, ce2 := err1.(gen.GeneratorError), err2.(gen.GeneratorError)
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
	ctx := gen.WithContext(context.Background(), gen.TestCtx{Writer: &b})
	err := g.Generate(ctx, testDoc, nil)
	if err == nil {
		t.Error(err)
		return
	}

	cerr, ok := err.(gen.GeneratorError)
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
	ctx := gen.WithContext(context.Background(), gen.TestCtx{Writer: &b})
	err := g.Generate(ctx, testDoc, nil)
	if err == nil {
		t.Error(err)
		return
	}

	cerr, ok := err.(gen.GeneratorError)
	if !ok {
		t.Fatal("unexpected err type")
		return
	}

	if cerr.Msg != "testing error response" {
		t.Fail()
	}
}

type testCtx struct {
	opener func(filename string) (io.WriteCloser, error)
	w      io.WriteCloser
}

func (ctx *testCtx) Open(filename string) (io.WriteCloser, error) {
	if ctx.opener != nil {
		return ctx.opener(filename)
	}
	return ctx.w, nil
}

type testErrWriter struct {
	err error
}

func (wc *testErrWriter) Write(b []byte) (int, error) { return 0, wc.err }
func (wc *testErrWriter) Close() error                { return wc.err }

func TestContextErrors(t *testing.T) {
	t.Run("ErrOnCtxOpen", func(subT *testing.T) {
		// Get helper cmd
		cmd := helperCommand(subT, "generate")

		// Create generate and run generate
		g := &Generator{
			Name: "test",
			Cmd:  cmd,
		}
		ctx := gen.WithContext(context.Background(), &testCtx{opener: func(string) (io.WriteCloser, error) { return nil, fmt.Errorf("test error") }})
		err := g.Generate(ctx, testDoc, map[string]interface{}{"hello": "world!"})
		if err == nil {
			subT.Errorf("expected error")
			return
		}

		if err.Error() != "compiler: generator error occurred in test:test test error" {
			subT.Error(err)
			return
		}
	})

	t.Run("ErrOnCtxWrite", func(subT *testing.T) {
		// Get helper cmd
		cmd := helperCommand(subT, "generate")

		// Create generate and run generate
		g := &Generator{
			Name: "test",
			Cmd:  cmd,
		}
		ctx := gen.WithContext(context.Background(), &testCtx{w: &testErrWriter{err: io.EOF}})
		err := g.Generate(ctx, testDoc, map[string]interface{}{"hello": "world!"})
		if err == nil {
			subT.Errorf("expected error")
			return
		}

		if err.Error() != "compiler: generator error occurred in test:test EOF" {
			subT.Error(err)
			return
		}
	})
}

// TestHelperProcess isn't a real test. It's used as a helper process
// for TestParameterRun.
//
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

		var req pb.Request
		err = proto.Unmarshal(b, &req)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}

		if len(req.FileToGenerate) != 1 {
			fmt.Fprintln(os.Stdout, "expected one file")
			os.Exit(0)
		}

		resp := &pb.Response{
			File: []*pb.Response_File{
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

		var req pb.Request
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

		var req pb.Request
		err = proto.Unmarshal(b, &req)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}

		if len(req.FileToGenerate) != 1 {
			fmt.Fprintln(os.Stdout, "expected one file")
			os.Exit(0)
		}

		resp := &pb.Response{
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
