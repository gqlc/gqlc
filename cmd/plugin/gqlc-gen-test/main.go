package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/gqlc/gqlc/cmd/plugin"
	"io/ioutil"
	"log"
	"os"
)

var testTxt = `Doc received: test, Opts: hello="world!"`

func main() {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	var req plugin.PluginRequest
	err = proto.Unmarshal(b, &req)
	if err != nil {
		log.Fatal(err)
	}

	if len(req.FileToGenerate) != 1 {
		log.Fatal("expected one file")
	}

	resp := &plugin.PluginResponse{
		File: []*plugin.PluginResponse_File{
			{
				Name:    "test.txt",
				Content: testTxt,
			},
		},
	}
	b, err = proto.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	_, err = os.Stdout.Write(b)
	if err != nil {
		log.Fatal(err)
	}
}
