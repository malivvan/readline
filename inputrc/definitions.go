package inputrc

const (
	emacs             = "emacs"
	hexValNum         = 10
	metaSeqLength     = 6
	storeDirectiveLen = 4
)

const (
	tokenNone token = iota
	tokenBind
	tokenBindMacro
	tokenSet
	tokenConstruct
)
