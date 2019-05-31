// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DONTCOVER

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// usage is a replacement usage function for the flags package.
func usage() {
	fmt.Fprintf(os.Stderr, "Usage of apidump:\n")
	fmt.Fprintf(os.Stderr, "\tapidump <pkg>\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("apidump: ")
	dirsInit()
	err := do(os.Stdout, flag.CommandLine, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}

// do is the workhorse, broken out of main to make testing easier.
func do(writer io.Writer, flagSet *flag.FlagSet, args []string) (err error) {
	flagSet.Usage = usage
	_ = flagSet.Parse(args)
	var paths []string
	var symbol, method string
	// Loop until something is printed.
	dirs.Reset()
	for i := 0; ; i++ {
		buildPackage, userPath, sym, more := parseArgs(flagSet.Args())
		if i > 0 && !more { // Ignore the "more" bit on the first iteration.
			return failMessage(paths, symbol, method)
		}
		if buildPackage == nil {
			return fmt.Errorf("no such package: %s", userPath)
		}
		symbol, method = parseSymbol(sym)
		pkg := parsePackage(writer, buildPackage, userPath)
		paths = append(paths, pkg.prettyPath())

		defer func() {
			pkg.flush()
			e := recover()
			if e == nil {
				return
			}
			pkgError, ok := e.(PackageError)
			if ok {
				err = pkgError
				return
			}
			panic(e)
		}()

		// We have a package.
		if symbol == "" {
			pkg.allDoc()
			return
		}
	}
}

// failMessage creates a nicely formatted error message when there is no result to show.
func failMessage(paths []string, symbol, method string) error {
	var b bytes.Buffer
	if len(paths) > 1 {
		b.WriteString("s")
	}
	b.WriteString(" ")
	for i, path := range paths {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(path)
	}
	if method == "" {
		return fmt.Errorf("no symbol %s in package%s", symbol, &b)
	}
	return fmt.Errorf("no method or field %s.%s in package%s", symbol, method, &b)
}

// parseArgs analyzes the arguments (if any) and returns the package
// it represents, the part of the argument the user used to identify
// the path (or "" if it's the current package) and the symbol
// (possibly with a .method) within that package.
// parseSymbol is used to analyze the symbol itself.
// The boolean final argument reports whether it is possible that
// there may be more directories worth looking at. It will only
// be true if the package path is a partial match for some directory
// and there may be more matches. For example, if the argument
// is rand.Float64, we must scan both crypto/rand and math/rand
// to find the symbol, and the first call will return crypto/rand, true.
func parseArgs(args []string) (pkg *build.Package, path, symbol string, more bool) {
	if len(args) == 0 {
		// Easy: current directory.
		return importDir(pwd()), "", "", false
	}
	arg := args[0]
	// We have an argument. If it is a directory name beginning with . or ..,
	// use the absolute path name. This discriminates "./errors" from "errors"
	// if the current directory contains a non-standard errors package.
	if isDotSlash(arg) {
		arg = filepath.Join(pwd(), arg)
	}
	switch len(args) {
	default:
		usage()
	case 1:
		// Done below.
	case 2:
		// Package must be findable and importable.
		pkg, err := build.Import(args[0], "", 0)
		if err == nil {
			return pkg, args[0], args[1], false
		}
		for {
			packagePath, ok := findNextPackage(arg)
			if !ok {
				break
			}
			if pkg, err := build.ImportDir(packagePath, 0); err == nil {
				return pkg, arg, args[1], true
			}
		}
		return nil, args[0], args[1], false
	}
	// Usual case: one argument.
	// If it contains slashes, it begins with a package path.
	// First, is it a complete package path as it is? If so, we are done.
	// This avoids confusion over package paths that have other
	// package paths as their prefix.
	pkg, err := build.Import(arg, "", 0)
	if err == nil {
		return pkg, arg, "", false
	}
	// Another disambiguator: If the symbol starts with an upper
	// case letter, it can only be a symbol in the current directory.
	// Kills the problem caused by case-insensitive file systems
	// matching an upper case name as a package name.
	if isUpper(arg) {
		pkg, err := build.ImportDir(".", 0)
		if err == nil {
			return pkg, "", arg, false
		}
	}
	// If it has a slash, it must be a package path but there is a symbol.
	// It's the last package path we care about.
	slash := strings.LastIndex(arg, "/")
	// There may be periods in the package path before or after the slash
	// and between a symbol and method.
	// Split the string at various periods to see what we find.
	// In general there may be ambiguities but this should almost always
	// work.
	var period int
	// slash+1: if there's no slash, the value is -1 and start is 0; otherwise
	// start is the byte after the slash.
	for start := slash + 1; start < len(arg); start = period + 1 {
		period = strings.Index(arg[start:], ".")
		symbol := ""
		if period < 0 {
			period = len(arg)
		} else {
			period += start
			symbol = arg[period+1:]
		}
		// Have we identified a package already?
		pkg, err := build.Import(arg[0:period], "", 0)
		if err == nil {
			return pkg, arg[0:period], symbol, false
		}
		// See if we have the basename or tail of a package, as in json for encoding/json
		// or ivy/value for robpike.io/ivy/value.
		pkgName := arg[:period]
		for {
			path, ok := findNextPackage(pkgName)
			if !ok {
				break
			}
			if pkg, err = build.ImportDir(path, 0); err == nil {
				return pkg, arg[0:period], symbol, true
			}
		}
		dirs.Reset() // Next iteration of for loop must scan all the directories again.
	}
	// If it has a slash, we've failed.
	if slash >= 0 {
		log.Fatalf("no such package %s", arg[0:period])
	}
	// Guess it's a symbol in the current directory.
	return importDir(pwd()), "", arg, false
}

// dotPaths lists all the dotted paths legal on Unix-like and
// Windows-like file systems. We check them all, as the chance
// of error is minute and even on Windows people will use ./
// sometimes.
var dotPaths = []string{
	`./`,
	`../`,
	`.\`,
	`..\`,
}

// isDotSlash reports whether the path begins with a reference
// to the local . or .. directory.
func isDotSlash(arg string) bool {
	if arg == "." || arg == ".." {
		return true
	}
	for _, dotPath := range dotPaths {
		if strings.HasPrefix(arg, dotPath) {
			return true
		}
	}
	return false
}

// importDir is just an error-catching wrapper for build.ImportDir.
func importDir(dir string) *build.Package {
	pkg, err := build.ImportDir(dir, 0)
	if err != nil {
		log.Fatal(err)
	}
	return pkg
}

// parseSymbol breaks str apart into a symbol and method.
// Both may be missing or the method may be missing.
// If present, each must be a valid Go identifier.
func parseSymbol(str string) (symbol, method string) {
	if str == "" {
		return
	}
	elem := strings.Split(str, ".")
	switch len(elem) {
	case 1:
	case 2:
		method = elem[1]
		isIdentifier(method)
	default:
		log.Printf("too many periods in symbol specification")
		usage()
	}
	symbol = elem[0]
	isIdentifier(symbol)
	return
}

// isIdentifier checks that the name is valid Go identifier, and
// logs and exits if it is not.
func isIdentifier(name string) {
	if len(name) == 0 {
		log.Fatal("empty symbol")
	}
	for i, ch := range name {
		if unicode.IsLetter(ch) || ch == '_' || i > 0 && unicode.IsDigit(ch) {
			continue
		}
		log.Fatalf("invalid identifier %q", name)
	}
}

// isExported reports whether the name is an exported identifier.
// If the unexported flag (-u) is true, isExported returns true because
// it means that we treat the name as if it is exported.
func isExported(name string) bool {
	return isUpper(name)
}

// isUpper reports whether the name starts with an upper case letter.
func isUpper(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(ch)
}

// findNextPackage returns the next full file name path that matches the
// (perhaps partial) package path pkg. The boolean reports if any match was found.
func findNextPackage(pkg string) (string, bool) {
	if pkg == "" || isUpper(pkg) { // Upper case symbol cannot be a package name.
		return "", false
	}
	if filepath.IsAbs(pkg) {
		if dirs.offset == 0 {
			dirs.offset = -1
			return pkg, true
		}
		return "", false
	}
	pkg = path.Clean(pkg)
	pkgSuffix := "/" + pkg
	for {
		d, ok := dirs.Next()
		if !ok {
			return "", false
		}
		if d.importPath == pkg || strings.HasSuffix(d.importPath, pkgSuffix) {
			return d.dir, true
		}
	}
}

var buildCtx = build.Default

// splitGopath splits $GOPATH into a list of roots.
func splitGopath() []string {
	return filepath.SplitList(buildCtx.GOPATH)
}

// pwd returns the current directory.
func pwd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
}
