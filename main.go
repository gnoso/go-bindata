// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	AppName    = "bindata"
	AppVersion = "1.0.0"
)

var (
	dir          = flag.String("dir", "", "Input directory. Processes all files in the path recursively.")
	out          = flag.String("o", "", "Path to the output files.")
	pkgname      = flag.String("p", "", "Name of the package to generate.")
	uncompressed = flag.Bool("u", false, "The specified resource will /not/ be GZIP compressed when this flag is specified. This alters the generated output code.")
	nomemcopy    = flag.Bool("m", false, "Use the memcopy hack to get rid of unnecessary memcopies. Refer to the documentation to see what implications this carries.")
)

func main() {
	flag.Parse()

	err := filepath.Walk(*dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && info.Name()[0] != '.' {

				inpath := strings.Replace(path, *dir+string(filepath.Separator), "", -1)
				funcName := genFunctionName(inpath)
				outfile := filepath.Join(*out, funcName+".go")

				// Only do translation if output file is out of date
				outFi, err := os.Stat(outfile)
				if err != nil || outFi.ModTime().Sub(info.ModTime()).Nanoseconds() < 0 {
					fmt.Fprintf(os.Stderr, "[i] translating file %s\n", path)

					err = translate_file(path, outfile, funcName, inpath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "[e] %s\n", err)
						return err
					}
				}
			}
			return nil
		})

	if err != nil {
		fmt.Fprintf(os.Stderr, "[e] %s\n", err)
		return
	}

	err = write_boilerplate(*out, *pkgname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[e] %s\n", err)
		return
	}
}

func write_boilerplate(outputDir, pkgName string) error {
	bp := fmt.Sprintf(`package %s
				var EmbeddedFiles map[string]func() []byte = make(map[string]func() []byte, 0)`, pkgName)

	if _, err := os.Stat(filepath.Join(outputDir, "boilerplate.go")); err != nil {
		return ioutil.WriteFile(filepath.Join(outputDir, "boilerplate.go"), []byte(bp), os.ModePerm)
	}
	return nil

}

func translate_file(infile string, outfile string, funcName string, inpath string) error {

	fs, err := os.Open(infile)
	if err != nil {
		return err
	}

	defer fs.Close()

	fd, err := os.Create(outfile)
	if err != nil {
		return err
	}

	defer fd.Close()

	translate(fs, fd, *pkgname, funcName, inpath, *uncompressed, *nomemcopy)
	return nil
}

func genFunctionName(filename string) string {
	funcName := filename
	funcName = strings.ToLower(funcName)
	funcName = strings.Replace(funcName, "/", "_", -1)
	funcName = strings.Replace(funcName, " ", "_", -1)
	funcName = strings.Replace(funcName, ".", "_", -1)
	funcName = strings.Replace(funcName, "-", "_", -1)

	if unicode.IsDigit(rune(funcName[0])) {
		// Identifier can't start with a digit.
		funcName = "_" + funcName
	}

	funcName = funcName
	return funcName
}
