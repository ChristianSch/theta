package outbound

// Used to add a post processor to the message pipeline
type PostProcessor struct {
	Processor LlmPostProcessor
	// Order is the order in which the post processor is applied, lower numbers are applied first
	Order int
	Name  string
}

type LlmPostProcessor interface {
	PostProcess(input []byte) ([]byte, error)
	GetName() string
}
