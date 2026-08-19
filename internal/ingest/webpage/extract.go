package webpage

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	readability "codeberg.org/readeck/go-readability/v2"
	nethtml "golang.org/x/net/html"
)

var ErrNoReadableContent = errors.New("webpage has no readable article content")

const readingWordsPerMinute = 250

func Extract(pageURL string, originalHTML []byte) (Document, error) {
	parsedURL, err := parseHTTPURL(pageURL)
	if err != nil {
		return Document{}, err
	}

	contentBaseURL := documentBaseURL(parsedURL, originalHTML)
	article, err := readability.FromReader(bytes.NewReader(originalHTML), contentBaseURL)
	if err != nil {
		return Document{}, fmt.Errorf("extract readable article: %w", err)
	}
	if article.Node == nil {
		return Document{}, ErrNoReadableContent
	}

	var cleanedHTML bytes.Buffer
	if err := article.RenderHTML(&cleanedHTML); err != nil {
		return Document{}, fmt.Errorf("render cleaned article HTML: %w", err)
	}
	var cleanedText bytes.Buffer
	if err := article.RenderText(&cleanedText); err != nil {
		return Document{}, fmt.Errorf("render cleaned article text: %w", err)
	}

	projector := blockProjector{
		pageURL: contentBaseURL,
		usedIDs: make(map[string]int),
	}
	projector.walkChildren(article.Node)
	if len(projector.blocks) == 0 || strings.TrimSpace(cleanedText.String()) == "" {
		return Document{}, ErrNoReadableContent
	}

	wordCount := countWords(cleanedText.String())
	document := Document{
		SourceURL:          parsedURL.String(),
		Title:              strings.TrimSpace(article.Title()),
		Byline:             strings.TrimSpace(article.Byline()),
		SiteName:           strings.TrimSpace(article.SiteName()),
		Language:           strings.TrimSpace(article.Language()),
		Excerpt:            strings.TrimSpace(article.Excerpt()),
		WordCount:          wordCount,
		ReadingTimeMinutes: readingTime(wordCount),
		LeadImageURL:       resolveURL(parsedURL, article.ImageURL()),
		ContentHTML:        cleanedHTML.String(),
		Blocks:             projector.blocks,
		Outline:            projector.outline,
	}
	if document.Title == "" {
		return Document{}, fmt.Errorf("%w: missing title", ErrNoReadableContent)
	}
	if document.SiteName == "" {
		document.SiteName = parsedURL.Hostname()
	}
	if publishedAt, err := article.PublishedTime(); err == nil && !publishedAt.IsZero() {
		document.PublishedAt = &publishedAt
	}
	if modifiedAt, err := article.ModifiedTime(); err == nil && !modifiedAt.IsZero() {
		document.ModifiedAt = &modifiedAt
	}
	document.Resources = collectResources(document)

	return document, nil
}

type blockProjector struct {
	pageURL *url.URL
	blocks  []Block
	outline []OutlineItem
	usedIDs map[string]int
}

func (p *blockProjector) walkChildren(node *nethtml.Node) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		p.walk(child)
	}
}

func (p *blockProjector) walk(node *nethtml.Node) {
	if node.Type != nethtml.ElementNode {
		return
	}

	switch strings.ToLower(node.Data) {
	case "article", "body", "div", "main", "section":
		p.walkChildren(node)
	case "h1", "h2", "h3", "h4", "h5", "h6":
		text := nodeText(node)
		if text == "" {
			return
		}
		level := int(node.Data[1] - '0')
		id := p.headingID(text)
		p.blocks = append(p.blocks, Block{Kind: BlockHeading, ID: id, Level: level, Text: text})
		p.outline = append(p.outline, OutlineItem{ID: id, Level: level, Title: text})
	case "p":
		images := descendantImages(node)
		for _, imageNode := range images {
			p.appendImage(imageNode, "")
		}
		text := nodeTextIgnoring(node, "img", "picture", "source")
		if text != "" {
			p.blocks = append(p.blocks, Block{Kind: BlockParagraph, Text: text, HTML: innerHTML(node)})
		}
	case "figure":
		caption := ""
		if captionNode := firstDescendant(node, "figcaption"); captionNode != nil {
			caption = nodeText(captionNode)
		}
		for _, imageNode := range descendantImages(node) {
			p.appendImage(imageNode, caption)
		}
	case "img":
		p.appendImage(node, "")
	case "ul", "ol":
		items := directChildren(node, "li")
		list := List{Ordered: node.Data == "ol", Items: make([]ListItem, 0, len(items))}
		for _, item := range items {
			if text := nodeText(item); text != "" {
				list.Items = append(list.Items, ListItem{Text: text, HTML: innerHTML(item)})
			}
		}
		if len(list.Items) > 0 {
			p.blocks = append(p.blocks, Block{Kind: BlockList, List: &list})
		}
	case "dl":
		definitions := projectDefinitions(node)
		if len(definitions) > 0 {
			p.blocks = append(p.blocks, Block{Kind: BlockDefinitions, Definitions: definitions})
		}
	case "blockquote":
		if text := nodeText(node); text != "" {
			p.blocks = append(p.blocks, Block{Kind: BlockQuote, Text: text, HTML: innerHTML(node)})
		}
	case "pre":
		if text := nodeTextPreservingWhitespace(node); text != "" {
			p.blocks = append(p.blocks, Block{Kind: BlockPreformatted, Text: text})
		}
	case "hr":
		p.blocks = append(p.blocks, Block{Kind: BlockDivider})
	case "table":
		p.blocks = append(p.blocks, Block{Kind: BlockHTML, HTML: renderNode(node)})
	case "script", "style", "template", "noscript", "figcaption", "source":
		return
	default:
		p.walkChildren(node)
	}
}

func (p *blockProjector) appendImage(node *nethtml.Node, caption string) {
	source := resolveURL(p.pageURL, attribute(node, "src"))
	if source == "" {
		return
	}
	p.blocks = append(p.blocks, Block{
		Kind: BlockImage,
		Image: &Image{
			URL:     source,
			Alt:     strings.TrimSpace(attribute(node, "alt")),
			Caption: caption,
		},
	})
}

func (p *blockProjector) headingID(text string) string {
	base := slug(text)
	p.usedIDs[base]++
	if p.usedIDs[base] == 1 {
		return base
	}
	return fmt.Sprintf("%s-%d", base, p.usedIDs[base])
}

func projectDefinitions(node *nethtml.Node) []Definition {
	var definitions []Definition
	var term string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != nethtml.ElementNode {
			continue
		}
		switch strings.ToLower(child.Data) {
		case "dt":
			term = nodeText(child)
		case "dd":
			description := nodeText(child)
			if term != "" && description != "" {
				definitions = append(definitions, Definition{Term: term, Description: description})
			}
		}
	}
	return definitions
}

func collectResources(document Document) []Resource {
	roles := make(map[string]string)
	var order []string
	add := func(rawURL, role string) {
		if rawURL == "" {
			return
		}
		if previous, ok := roles[rawURL]; ok {
			if role == "lead-image" && previous != role {
				roles[rawURL] = role
			}
			return
		}
		roles[rawURL] = role
		order = append(order, rawURL)
	}
	add(document.LeadImageURL, "lead-image")
	for _, block := range document.Blocks {
		if block.Image != nil {
			add(block.Image.URL, "content-image")
		}
	}
	resources := make([]Resource, 0, len(order))
	for _, resourceURL := range order {
		resources = append(resources, Resource{URL: resourceURL, Role: roles[resourceURL]})
	}
	return resources
}

func parseHTTPURL(rawURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse webpage URL: %w", err)
	}
	if (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		return nil, fmt.Errorf("webpage URL must be absolute HTTP or HTTPS: %q", rawURL)
	}
	return parsedURL, nil
}

func documentBaseURL(pageURL *url.URL, originalHTML []byte) *url.URL {
	document, err := nethtml.Parse(bytes.NewReader(originalHTML))
	if err != nil {
		return pageURL
	}
	var findBase func(*nethtml.Node) string
	findBase = func(node *nethtml.Node) string {
		if node.Type == nethtml.ElementNode && strings.EqualFold(node.Data, "base") {
			return attribute(node, "href")
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if href := findBase(child); href != "" {
				return href
			}
		}
		return ""
	}
	href := findBase(document)
	if href == "" {
		return pageURL
	}
	base, err := url.Parse(href)
	if err != nil {
		return pageURL
	}
	resolved := pageURL.ResolveReference(base)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return pageURL
	}
	return resolved
}

func resolveURL(base *url.URL, reference string) string {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return ""
	}
	parsed, err := url.Parse(reference)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(parsed)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	return resolved.String()
}

func attribute(node *nethtml.Node, name string) string {
	for _, attribute := range node.Attr {
		if strings.EqualFold(attribute.Key, name) {
			return attribute.Val
		}
	}
	return ""
}

func descendantImages(node *nethtml.Node) []*nethtml.Node {
	var result []*nethtml.Node
	var visit func(*nethtml.Node)
	visit = func(current *nethtml.Node) {
		if current.Type == nethtml.ElementNode && strings.EqualFold(current.Data, "img") {
			result = append(result, current)
			return
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(node)
	return result
}

func firstDescendant(node *nethtml.Node, tag string) *nethtml.Node {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == nethtml.ElementNode && strings.EqualFold(child.Data, tag) {
			return child
		}
		if match := firstDescendant(child, tag); match != nil {
			return match
		}
	}
	return nil
}

func directChildren(node *nethtml.Node, tag string) []*nethtml.Node {
	var result []*nethtml.Node
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == nethtml.ElementNode && strings.EqualFold(child.Data, tag) {
			result = append(result, child)
		}
	}
	return result
}

func nodeText(node *nethtml.Node) string {
	var text strings.Builder
	collectText(node, &text, nil)
	return normalizeSpaces(text.String())
}

func nodeTextIgnoring(node *nethtml.Node, ignoredTags ...string) string {
	ignored := make(map[string]bool, len(ignoredTags))
	for _, tag := range ignoredTags {
		ignored[strings.ToLower(tag)] = true
	}
	var text strings.Builder
	collectText(node, &text, ignored)
	return normalizeSpaces(text.String())
}

func collectText(node *nethtml.Node, text *strings.Builder, ignored map[string]bool) {
	if node.Type == nethtml.ElementNode && ignored[strings.ToLower(node.Data)] {
		return
	}
	if node.Type == nethtml.TextNode {
		text.WriteString(node.Data)
		text.WriteByte(' ')
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		collectText(child, text, ignored)
	}
}

func nodeTextPreservingWhitespace(node *nethtml.Node) string {
	var text strings.Builder
	var visit func(*nethtml.Node)
	visit = func(current *nethtml.Node) {
		if current.Type == nethtml.TextNode {
			text.WriteString(current.Data)
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(node)
	return strings.TrimSpace(text.String())
}

func normalizeSpaces(value string) string {
	return strings.Join(strings.Fields(html.UnescapeString(value)), " ")
}

func innerHTML(node *nethtml.Node) string {
	var output strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if err := nethtml.Render(&output, child); err != nil {
			return ""
		}
	}
	return output.String()
}

func renderNode(node *nethtml.Node) string {
	var output strings.Builder
	if err := nethtml.Render(&output, node); err != nil {
		return ""
	}
	return output.String()
}

func countWords(value string) int {
	count := 0
	inWord := false
	for len(value) > 0 {
		r, size := utf8.DecodeRuneInString(value)
		value = value[size:]
		isWordRune := unicode.IsLetter(r) || unicode.IsNumber(r)
		if isWordRune && !inWord {
			count++
		}
		inWord = isWordRune
	}
	return count
}

func readingTime(words int) int {
	if words <= 0 {
		return 0
	}
	return (words + readingWordsPerMinute - 1) / readingWordsPerMinute
}

func slug(value string) string {
	var output strings.Builder
	separator := false
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if separator && output.Len() > 0 {
				output.WriteByte('-')
			}
			output.WriteRune(r)
			separator = false
		} else {
			separator = true
		}
	}
	result := strings.Trim(output.String(), "-")
	if result == "" {
		return "section"
	}
	return result
}
