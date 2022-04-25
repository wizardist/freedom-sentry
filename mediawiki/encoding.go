package mediawiki

const listSeparator = "|"

type TextBool string

const (
	TextBoolYes      TextBool = "yes"
	TextBoolNo       TextBool = "no"
	TextBoolNoChange TextBool = "nochange"
)
