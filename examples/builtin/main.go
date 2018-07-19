package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/liuzl/fmr"
	"github.com/liuzl/goutil"
	"github.com/robertkrimen/otto"
)

var (
	grammar = flag.String("g", "builtin.grammar", "grammar file")
	js      = flag.String("js", "math.js", "javascript file")
	input   = flag.String("i", "", "file of original text to read")
	debug   = flag.Bool("debug", false, "debug mode")
	start   = flag.String("start", "number", "start rule")
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	flag.Parse()
	b, err := ioutil.ReadFile(*grammar)
	if err != nil {
		glog.Fatal(err)
	}
	if *debug {
		fmr.Debug = true
	}
	g, err := fmr.CFGrammar(string(b))
	if err != nil {
		glog.Fatal(err)
	}
	if *debug {
		b, err = goutil.JsonMarshalIndent(g, "", "    ")
		if err != nil {
			glog.Fatal(err)
		}
		fmt.Printf("%s\n", string(b))
	}
	script, err := ioutil.ReadFile(*js)
	if err != nil {
		glog.Fatal(err)
	}
	vm := otto.New()
	if _, err = vm.Run(script); err != nil {
		glog.Fatal(err)
	}

	var in *os.File
	if *input == "" {
		in = os.Stdin
	} else {
		in, err = os.Open(*input)
		if err != nil {
			glog.Fatal(err)
		}
		defer in.Close()
	}
	br := bufio.NewReader(in)

	for {
		line, c := br.ReadString('\n')
		if c == io.EOF {
			break
		}
		if c != nil {
			glog.Fatal(c)
		}
		line = strings.TrimSpace(line)
		fmt.Println(line)
		//p, err := g.EarleyParse(*start, line)
		//fmt.Println(p, err)
		//TODO p.GetTrees for (any)
		//p.GetTrees()

		ps, err := g.EarleyParseAll(line, *start)
		if err != nil {
			glog.Fatal(err)
		}
		for i, p := range ps {
			trees := p.GetTrees()
			//fmt.Printf("%+v\n", p)
			fmt.Printf("p%d tree number:%d\n", i, len(trees))
			for _, tree := range trees {
				//tree.Print(os.Stdout)
				sem, err := tree.Semantic()
				//fmt.Println(sem)
				if err != nil {
					glog.Fatal(err)
				}
				if *debug {
					fmt.Printf("%s = ?\n", sem)
				}
				result, err := vm.Run(sem)
				if err != nil {
					glog.Error(err)
				}
				rs, _ := result.Export()
				fmt.Printf("%s = %+v\n", sem, rs)
				//eval, err := tree.Eval()
				//fmt.Printf("Eval: %s, Err: %+v\n", eval, err)
			}
		}
		fmt.Println()
	}
}
