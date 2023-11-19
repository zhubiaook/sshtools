package options

type Options struct {
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
}

func NewOptions() Options {
	return Options{}
}
