package gen

// Option is a function that configures generation parameters.
type Option func(*Params)

// Params holds configuration options for generation.
type Params struct {
	Temperature *float64
	JSONMode    *bool
}

// WithTemperature sets the temperature parameter for generation.
func WithTemperature(temp float64) Option {
	return func(p *Params) {
		p.Temperature = &temp
	}
}

// WithJSONMode enables JSON mode for generation.
func WithJSONMode() Option {
	return func(p *Params) {
		jm := true
		p.JSONMode = &jm
	}
}
