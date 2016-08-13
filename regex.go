package main

import (
	"regexp"
)

var (
	reGetPID = regexp.MustCompile(`(\d+) .*`)

	reExtract = regexp.MustCompile(`(\d+\.\d+) (\w+)(\(.*) <(.+)>`)

	reExtractNoElapsed = regexp.MustCompile(`(\d+\.\d+) (\w+)(\(.*)`)

	reExtractUnfinished = regexp.MustCompile(`\s*(\d+\.\d+ .*) <unfinished \.\.\.>$`)

	reExtractResumed = regexp.MustCompile(`\s*(\d+\.\d+) <\.\.\. [\a-zA-Z\d]+ resumed>(.*)$`)

	reExtractSignal = regexp.MustCompile(`\s*(\d+\.\d+) --- (\w+) \(([\w ]+)\) @ (\d)+ \((\d+)\) ---$`)

	reExtractArgumentsAndReturnValueNone = regexp.MustCompile(`\((.*)\)[ \t]*= (\?)$`)

	reExtractArgumentsAndReturnValueOk = regexp.MustCompile(`\((.*)\)[ \t]*= (-?\d+)$`)

	reExtractArgumentsAndReturnValueOkHex = regexp.MustCompile(`\((.*)\)[ \t]*= (-?0[xX][a-fA-F\d]+)$`)

	reExtractArgumentsAndReturnValueError = regexp.MustCompile(`\((.*)\)[ \t]*= (-?\d+) (\w+) \([\w ]+\)$`)

	reExtractArgumentsAndReturnValueErrorUnknown = regexp.MustCompile(`\((.*)\)[ \t]*= (\?) (\w+) \([\w ]+\)$`)

	reExtractArgumentsAndReturnValueExt = regexp.MustCompile(`\((.*)\)[ \t]*= (-?\d+) \(([^()]+)\)$`)

	reExtractArgumentsAndReturnValueExtHex = regexp.MustCompile(`\((.*)\)[ \t]*= (-?0[xX][a-fA-F\d]+) \(([^()]+)\)$`)

	reExtractArgs = regexp.MustCompile(`\(([^\)]+)\)`)

	reExtractReturnValue = regexp.MustCompile(`\=(.*)`)
)
