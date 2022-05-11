package graphviz

import (
	"fmt"
	"strings"
)

type Attributes struct {
	attrs map[string]string
}

func NewAttributes() *Attributes {
	return &Attributes{attrs: map[string]string{}}
}

func (a *Attributes) SetAttr(name, value string) { a.attrs[name] = value }

func (n *Attributes) SetShape(shape string) { n.SetAttr("shape", shape) }

func (n *Attributes) SetColor(color string) { n.SetAttr("color", color) }

func (n *Attributes) SetBgColor(color string) { n.SetAttr("bgcolor", color) }

func (n *Attributes) SetComment(comment string) { n.SetAttr("comment", comment) }

func (n *Attributes) SetLabel(label string) { n.SetAttr("label", label) }

func (n *Attributes) SetPenWidth(w string) { n.SetAttr("penwidth", w) }

func (n *Attributes) SetFontColor(color string) { n.SetAttr("fontcolor", color) }

func (a *Attributes) String() string {
	if len(a.attrs) == 0 {
		return ""
	}
	var attrStrs []string
	for k, v := range a.attrs {
		attrStrs = append(attrStrs, fmt.Sprintf("%s=%q", k, v))
	}
	return fmt.Sprintf("[%s]", strings.Join(attrStrs, ", "))
}
