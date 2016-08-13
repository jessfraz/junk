package main

import (
	"regexp"
)

var (
	ReGetPID = regexp.MustCompile(`(\d+) .*`)

	ReExtract = regexp.MustCompile(`(\d+\.\d+) (\w+)(\(.*) <(.+)>`)

	ReExtractNoElapsed = regexp.MustCompile(`(\d+\.\d+) (\w+)(\(.*)`)

	ReExtractUnfinished = regexp.MustCompile(`\s*(\d+\.\d+ .*) <unfinished \.\.\.>$`)

	ReExtractResumed = regexp.MustCompile(`\s*(\d+\.\d+) <\.\.\. [\a-zA-Z\d]+ resumed>(.*)$`)

	ReExtractSignal = regexp.MustCompile(`\s*(\d+\.\d+) --- (\w+) \(([\w ]+)\) @ (\d)+ \((\d+)\) ---$`)

	ReExtractArgumentsAndReturnValueNone = regexp.MustCompile(`\((.*)\)[ \t]*= (\?)$`)

	ReExtractArgumentsAndReturnValueOk = regexp.MustCompile(`\((.*)\)[ \t]*= (-?\d+)$`)

	ReExtractArgumentsAndReturnValueOkHex = regexp.MustCompile(`\((.*)\)[ \t]*= (-?0[xX][a-fA-F\d]+)$`)

	ReExtractArgumentsAndReturnValueError = regexp.MustCompile(`\((.*)\)[ \t]*= (-?\d+) (\w+) \([\w ]+\)$`)

	ReExtractArgumentsAndReturnValueErrorUnknown = regexp.MustCompile(`\((.*)\)[ \t]*= (\?) (\w+) \([\w ]+\)$`)

	ReExtractArgumentsAndReturnValueExt = regexp.MustCompile(`\((.*)\)[ \t]*= (-?\d+) \(([^()]+)\)$`)

	ReExtractArgumentsAndReturnValueExtHex = regexp.MustCompile(`\((.*)\)[ \t]*= (-?0[xX][a-fA-F\d]+) \(([^()]+)\)$`)

	ReExtractArgs = regexp.MustCompile(`\(([^\)]+)\)`)

	ReExtractReturnValue = regexp.MustCompile(`\=(.*)`)
)
