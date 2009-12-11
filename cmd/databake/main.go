// Code generator for converting binary files into go source files.

package main

import (
	"flag";
	"fmt";
	"io/ioutil";
	"os";
	"regexp";
)

var (
	packageName = flag.String("package", "",
		"package name in generated source; omit package declaration if empty");
	variableName = flag.String("variable", "Data",
		"data variable name in generated source");
)

func usage() {
	fmt.Fprintf(os.Stderr,
		"usage: databake [OPTION]... [SOURCE] [DEST]\n"
		"       If SOURCE or DEST is omitted or '-', use stdin / stdout.\n");
	flag.PrintDefaults();
	os.Exit(2);
}

const pipeName = "-"

func main() {
	flag.Usage = usage;
	flag.Parse();

	if !isSymbolName(*variableName) {
		die(fmt.Sprintf(
			"Variable name '%v' is not a valid Go identifier.",
			*variableName));
	}

	input := os.Stdin;
	output := os.Stdout;

	// Get input file.
	if flag.NArg() > 0 && flag.Arg(0) != pipeName {
		var err os.Error;

 		input, err = os.Open(flag.Arg(0), os.O_RDONLY, 0);
		defer input.Close();
		dieIfErr("Input file error:", err);
	}

	// Get output file
	if flag.NArg() > 1 && flag.Arg(1) != pipeName {
		var err os.Error;

		// Open file with read and write permissions for owner and
		// read permission for others.
		output, err = os.Open(flag.Arg(1), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644);
		defer output.Close();
		dieIfErr("Output file error:", err);
	}

	if *packageName != "" {
		fmt.Fprintf(output, "package %v\n\n", *packageName);
	}


	data, err := ioutil.ReadAll(input);
	dieIfErr("Read error:", err);

	if len(data) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: No data found.\n");
	}

	// XXX: As of 2009-12-05, emitting a byte array literal seems to
	// produce a *much* larger object (x22) code block than the number of
	// bytes in the data. So it's not really usable for embedding
	// purposes. The current implementation uses a string constant
	// instead.

	// printByteArray(output, *variableName, data);
	printConstString(output, *variableName, data);
}

func printByteArray(output *os.File, variableName string, data []byte) {
	fmt.Fprintf(output, "var %v = [...]byte{\n", variableName);

	const bytesPerLine = 16;

	for i := 0; i < len(data); i += bytesPerLine {
		for j := 0; j < bytesPerLine && i + j < len(data); j++ {
			fmt.Fprintf(output, "0x%02x,", data[i + j]);
		}
		fmt.Fprintf(output, "\n");
	}
	fmt.Fprintf(output, "}\n");
}

func printConstString(output *os.File, variableName string, data []byte) {
	fmt.Fprintf(output, "const %v =\n", variableName);

	const bytesPerLine = 16;
	const printCleartext = false;

	for i := 0; i < len(data); i += bytesPerLine {
		fmt.Fprintf(output, "\"");
		for j := 0; j < bytesPerLine && i + j < len(data); j++ {
			byte := data[i + j];
			if printCleartext && byte >= ' ' && byte <= '~' {
				fmt.Fprintf(output, "%c", byte);
			} else {
				fmt.Fprintf(output, "\\x%02x", byte);
			}
		}
		fmt.Fprintf(output, "\"\n");
	}
}

func isSymbolName(name string) bool {
	// XXX: Regexp doesn't handle unicode.

	// XXX: Could this use the Go language parsing library?

	// XXX: Doesn't check if the name is a reserved word.
	const symbolRE = "[_A-Za-z][_A-Za-z0-9]*";
	matched, err := regexp.MatchString(symbolRE, name);
	dieIfErr("Regexp match error:", err);
	return matched;
}

func dieIfErr(msg string, err fmt.Stringer) {
	if err != nil {
		die(fmt.Sprintf("%v %v\n", msg, err));
	}
}

func die(msg string) {
	fmt.Fprintf(os.Stderr, "%v\n", msg);
	os.Exit(2);
}