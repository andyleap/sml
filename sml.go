package sml

var (
	g = makeGrammar()
)

func Decode(rs RuneSeeker) (interface{}, error) {
	isr := &IndentStateReader{rs: rs}
	return g(isr)
}
