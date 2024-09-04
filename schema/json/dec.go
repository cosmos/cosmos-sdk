package schemajson

type parsedDecimal struct {
	negative        bool
	digitsBeforeDot []byte
	digitsAfterDot  []byte
	exponentDigits  []byte
}

func parseDecimalString(str string) {

}
