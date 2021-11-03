package rs

var (
	_ TypeHint = TypeHintNone{}
	_ TypeHint = TypeHintStr{}
	_ TypeHint = TypeHintObject{}
	_ TypeHint = TypeHintObjects{}
	_ TypeHint = TypeHintInt{}
	_ TypeHint = TypeHintFloat{}
	_ TypeHint = TypeHintBool{}
)
