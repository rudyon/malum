package document

import (
	"fmt"
	"strings"

	nethtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func rewriteResourcePaths(content string, storedPaths map[string]string) ([]byte, error) {
	context := &nethtml.Node{
		Type:     nethtml.ElementNode,
		DataAtom: atom.Div,
		Data:     "div",
	}
	nodes, err := nethtml.ParseFragment(strings.NewReader(content), context)
	if err != nil {
		return nil, fmt.Errorf("parse normalized article HTML: %w", err)
	}

	for _, node := range nodes {
		rewriteNodeResourcePaths(node, storedPaths)
	}

	var output strings.Builder
	for _, node := range nodes {
		if err := nethtml.Render(&output, node); err != nil {
			return nil, fmt.Errorf("render normalized article HTML: %w", err)
		}
	}
	return []byte(output.String()), nil
}

func rewriteNodeResourcePaths(node *nethtml.Node, storedPaths map[string]string) {
	if node.Type == nethtml.ElementNode {
		for index := range node.Attr {
			attribute := &node.Attr[index]
			if strings.EqualFold(attribute.Key, "src") {
				if localPath, ok := storedPaths[attribute.Val]; ok {
					attribute.Val = localPath
				}
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		rewriteNodeResourcePaths(child, storedPaths)
	}
}
