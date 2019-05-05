package plugin

import (
	"bytes"
	"context"
	"github.com/gqlc/compiler"
	"github.com/gqlc/graphql/parser"
	"github.com/gqlc/graphql/token"
	"github.com/spf13/afero"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	goCmd := exec.Command("go", "install", "github.com/gqlc/gqlc/cmd/plugin/gqlc-gen-test")
	err := goCmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

var (
	testGql = `scalar Test`

	testTxt = `Doc received: test, Opts: hello="world!"`
)

type genCtx struct {
	fs  afero.Fs
	dir string
}

func (ctx *genCtx) Open(name string) (io.WriteCloser, error) {
	return ctx.fs.OpenFile(filepath.Join(ctx.dir, name), os.O_WRONLY|os.O_CREATE, 0755)
}

func TestGenerator_Generate(t *testing.T) {
	// Parse test doc
	doc, err := parser.ParseDoc(token.NewDocSet(), "test", strings.NewReader(testGql), 0)
	if err != nil {
		t.Error(err)
		return
	}

	// Set up test file system
	fs := afero.NewMemMapFs()

	// Create generate and run generate
	g := &Generator{
		Name:   "test",
		Prefix: "gqlc-gen-",
	}
	ctx := compiler.WithContext(context.Background(), &genCtx{fs: fs, dir: "/home"})
	err = g.Generate(ctx, doc, `{hello: "world!"}`)
	if err != nil {
		t.Error(err)
		return
	}

	// Check test file was written
	f, err := fs.Open("/home/test.txt")
	if err != nil {
		t.Error(err)
		return
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.EqualFold(b, []byte(testTxt)) {
		t.Fail()
		return
	}
}
