package feedxml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

// Parse reads a whole feed document into a Node tree.
// thin wrapper around ParseBytes
func Parse(reader io.Reader) (*Node, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("feedxml.Parse: %w", err)
	}
	return ParseBytes(raw)
}

// ParseBytes reads a whole feed document into a Node tree.
func ParseBytes(document []byte) (*Node, error) {
	decoder := xml.NewDecoder(bytes.NewReader(sanitizeXML(decodeUTF8(document))))
	decoder.Strict = false
	decoder.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) {
		return input, nil
	}

	var root *Node
	var stack []*Node
	// texts[i] accumulates the character data of stack[i]. Appending here
	// rather than to Node.Text keeps accumulation linear in token count.
	var texts [][]byte

	flush := func() {
		last := len(stack) - 1
		stack[last].Text = string(texts[last])
		stack, texts = stack[:last], texts[:last]
	}

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			for len(stack) > 0 {
				flush()
			}
			return root, fmt.Errorf("feedxml.ParseBytes: %w", err)
		}

		switch token := token.(type) {
		case xml.StartElement:
			node := &Node{
				Space: CanonNS(token.Name.Space),
				Local: token.Name.Local,
				Attrs: make(map[string]string, len(token.Attr)),
			}
			// Attrs is keyed by local name, so a namespaced x:url and a plain
			// url collide. The plain one wins whatever the document order.
			for _, attribute := range token.Attr {
				if attribute.Name.Local == "xmlns" || attribute.Name.Space == "xmlns" {
					continue
				}
				if attribute.Name.Space != "" {
					if _, taken := node.Attrs[attribute.Name.Local]; taken {
						continue
					}
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
			texts = append(texts, nil)
		case xml.EndElement:
			if len(stack) > 0 {
				flush()
			}

		case xml.CharData:
			if len(stack) > 0 {
				last := len(texts) - 1
				texts[last] = append(texts[last], token...)
			}
		}
	}

	// Elements left open at EOF still own their text.
	for len(stack) > 0 {
		flush()
	}

	if root == nil {
		return nil, errors.New("feedxml.ParseBytes: no root element")
	}
	return root, nil
}
