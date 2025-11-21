package app

type SpecValidationError struct {
	Msg string
}

func (e *SpecValidationError) Error() string {
	return "spec validation error: " + e.Msg
}

type MaterializationError struct {
	Msg string
}

func (e *MaterializationError) Error() string {
	return "materialization error: " + e.Msg
}
