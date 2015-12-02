# Simple Go Regex Matcher

*Very* simple, but fairly quick, regular expression matching program written in Go for a school project. Only supports ( ), |, +, *, {n}.

This program uses Thompson's NFA algorithm, and is based in part on the implementation and explanation by [Russ Cox][swtch].

## How to compile
The program is quite flexible and can be built and ran in a number of ways.

By far the easiest is to use the mostly-up-to-date version on the Go Playground here: https://play.golang.org/p/JMkw-zeCYI. Just scroll to the end of the file and replace the placeholders with your regex and string and click Run (or click Run first and the line numbers of the string and regex are displayed).

If you want to compile on your own without setting up GOPATHs and whatnot, run `go run theoryregex.go` in whatever folder theoryregex.go is in. Or run the outputted executable from `go build theoryregex.go`.

There are also pre-compiled binaries available.

## How to use

When running the program outside of the Go Playground, the program can be run either interactively or by passing parameters. If the program is run with any number of parameters other than 2, it will prompt for a regular expression and a string to test interactively. Alternatively, the program can be run in the format of `./theoryregex regex string`. In either case, the program will print `true` or `false` depending on if the regular expression can build the string. The program exits with status 0 with a true result and 1 with a false result.

[swtch]: https://swtch.com/~rsc/regexp/regexp1.html
