package chain

type OptionChainProcessor interface {
	Name() string
	Analyze(optionChains *Chains) error
}
