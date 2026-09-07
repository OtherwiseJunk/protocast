package feedxml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	"golang.org/x/net/html/charset"
)

// Parse reads a whole feed document into a Node tree
func Parse(reader io.Reader) (*Node, error) {
	decoder := xml.NewDecoder(reader)
	decoder.Strict = false
	decoder.CharsetReader = charset.NewReaderLabel

	var root *Node
	var stack []*Node

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("feedxml.Parse: %w", err)
		}

		switch token := token.(type) {
		case xml.StartElement:
			node := &Node{
				Space: CanonNS(token.Name.Space),
				Local: token.Name.Local,
				Attrs: make(map[string]string, len(token.Attr)),
			}
			for _, attribute := range token.Attr {
				if attribute.Name.Local == "xmlns" || attribute.Name.Space == "xmlns" {
					continue
				}
				node.Attrs[attribute.Name.Local] = attribute.Value
			}
			if len(stack) == 0 {
				if root == nil {
					root = node
				}
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].Text += string(token)
			}
		}
	}

	if root == nil {
		return nil, errors.New("feedxml.Parse: no root element")
	}
	return root, nil
}
