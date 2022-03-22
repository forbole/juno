package pricefeed

type Config struct {
	Tokens []Token `yaml:"tokens"`
}

type Token struct {
	Name  string `yaml:"name"`
	Units []Coin `yaml:"units"`
}

type Coin struct {
	Denom    string `yaml:"denom"`
	Exponent int    `yaml:"exponent"`
	PriceID  string `yaml:"price_id,omitempty"`
}

// NewPricefeedConfig allows to build a new Config instance
func NewPricefeedConfig(tokens []Token) Config {
	return Config{Tokens: tokens}
}

// DefaultPricefeedConfig returns the default instance of Config
func DefaultPricefeedConfig() Config {
	token := Token{
		Name: "desmos",
		Units: []Coin{
			{
				Denom:    "udesmos",
				Exponent: 0,
			},
			{
				Denom:    "desmos",
				Exponent: 6,
				PriceID:  "desmos",
			},
		},
	}
	return NewPricefeedConfig([]Token{token})
}
