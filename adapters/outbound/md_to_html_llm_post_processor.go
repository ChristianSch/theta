package outbound

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type MdToHtmlLlmPostProcessor struct {
}

func NewMdToHtmlLlmPostProcessor() *MdToHtmlLlmPostProcessor {
	return &MdToHtmlLlmPostProcessor{}
}

func (pp *MdToHtmlLlmPostProcessor) PostProcess(input []byte) ([]byte, error) {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(input)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer), nil
}

func (pp *MdToHtmlLlmPostProcessor) GetName() string {
	return "md_to_html_llm_post_processor"
}
