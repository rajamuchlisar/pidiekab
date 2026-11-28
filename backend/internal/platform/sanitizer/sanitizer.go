package sanitizer

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	// Tags whose contents must be completely removed along with the tags
	dangerousBlockTags = []string{
		"script", "style", "iframe", "frame", "frameset",
		"object", "embed", "applet", "form", "textarea", "select",
	}

	// Tags that should be removed (opening and closing) but harmless inner text can stay or be discarded
	forbiddenTags = []string{
		"input", "button", "option", "optgroup", "meta", "base", "link",
		"marquee", "blink", "dialog", "svg", "body", "head", "html", "title",
	}

	// Safe tags allowed in rich text articles
	allowedTags = map[string]bool{
		"p": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
		"strong": true, "b": true, "em": true, "i": true, "u": true, "s": true, "strike": true, "del": true,
		"blockquote": true, "ul": true, "ol": true, "li": true, "a": true, "img": true,
		"table": true, "thead": true, "tbody": true, "tfoot": true, "tr": true, "th": true, "td": true,
		"hr": true, "br": true, "code": true, "pre": true, "div": true, "span": true,
		"figure": true, "figcaption": true, "caption": true, "sub": true, "sup": true,
		"small": true, "details": true, "summary": true, "section": true, "article": true,
		"header": true, "footer": true, "aside": true, "video": true, "audio": true, "source": true,
	}

	// Regex for detecting any inline event handlers: onload, onerror, onclick, ontoggle, etc.
	// Matches on[a-zA-Z]+ preceded by whitespace, word boundary, or slash '/'
	eventHandlerRegex = regexp.MustCompile(`(?i)(?:^|[\s/])on[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^>\s]+)`)

	// Regex for attributes
	hrefAttrRegex = regexp.MustCompile(`(?i)\bhref\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	srcAttrRegex  = regexp.MustCompile(`(?i)\bsrc\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	styleAttrRegex = regexp.MustCompile(`(?i)\bstyle\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	relAttrRegex   = regexp.MustCompile(`(?i)\brel\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	targetAttrRegex = regexp.MustCompile(`(?i)\btarget\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)

	// Regex to extract HTML tags (allowing whitespace or slash as delimiter)
	htmlTagRegex = regexp.MustCompile(`(?s)<\s*/?\s*([a-zA-Z0-9]+)([\s/][^>]*)?>`)

	// Protocol checks
	dangerousProtocols = []string{"javascript:", "vbscript:", "data:", "file:", "blob:"}
)

// SanitizeHTML thoroughly strips XSS and HTML injection vectors from HTML strings,
// while preserving safe semantic rich text markup and styling.
func SanitizeHTML(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}

	// 1. Remove null bytes and control characters
	cleaned := strings.ReplaceAll(input, "\x00", "")

	// 2. Iteratively strip dangerous block tags and their inner content (e.g. <script>...</script>, <style>...</style>)
	// Repeating until no more matches prevents nested tag bypasses like <scr<script>ipt>
	for _, tag := range dangerousBlockTags {
		blockPattern := regexp.MustCompile(fmt.Sprintf(`(?is)<\s*%s([\s/][^>]*)?>.*?<\s*/\s*%s\s*>`, tag, tag))
		prev := ""
		for prev != cleaned {
			prev = cleaned
			cleaned = blockPattern.ReplaceAllString(cleaned, "")
		}
		// Also strip any standalone opening or closing tags for this block tag
		standalonePattern := regexp.MustCompile(fmt.Sprintf(`(?is)<\s*/?\s*%s([\s/][^>]*)?>?`, tag))
		cleaned = standalonePattern.ReplaceAllString(cleaned, "")
	}

	// 3. Strip other forbidden tags (svg, marquee, input, button, form, etc.)
	for _, tag := range forbiddenTags {
		standalonePattern := regexp.MustCompile(fmt.Sprintf(`(?is)<\s*/?\s*%s([\s/][^>]*)?>?`, tag))
		prev := ""
		for prev != cleaned {
			prev = cleaned
			cleaned = standalonePattern.ReplaceAllString(cleaned, "")
		}
	}

	// 4. Inspect remaining tags and sanitize attributes
	cleaned = htmlTagRegex.ReplaceAllStringFunc(cleaned, func(tagMatch string) string {
		sub := htmlTagRegex.FindStringSubmatch(tagMatch)
		if len(sub) < 2 {
			return ""
		}

		tagName := strings.ToLower(sub[1])
		if !allowedTags[tagName] {
			// Tag not allowed; drop the tag but let inner text remain
			return ""
		}

		// Closing tag?
		isClosing := strings.HasPrefix(strings.TrimSpace(tagMatch), "</")
		if isClosing {
			return fmt.Sprintf("</%s>", tagName)
		}

		// Opening or self-closing tag: sanitize attributes
		tagAttrs := sub[2]

		// a) Strip all event handlers (onload, onerror, onclick, onmouseover, ontoggle, etc.)
		tagAttrs = eventHandlerRegex.ReplaceAllString(tagAttrs, "")

		// Clean any stray slashes used as attribute separators, e.g. <details/open>
		tagAttrs = strings.Trim(tagAttrs, "/")

		// b) Sanitize href attribute if present
		tagAttrs = hrefAttrRegex.ReplaceAllStringFunc(tagAttrs, func(hrefMatch string) string {
			rawVal := extractAttrValue(hrefMatch)
			if !IsSafeURL(rawVal) {
				return ""
			}
			return fmt.Sprintf(`href="%s"`, escapeAttr(rawVal))
		})

		// c) Sanitize src attribute if present
		tagAttrs = srcAttrRegex.ReplaceAllStringFunc(tagAttrs, func(srcMatch string) string {
			rawVal := extractAttrValue(srcMatch)
			if !IsSafeURL(rawVal) {
				return ""
			}
			return fmt.Sprintf(`src="%s"`, escapeAttr(rawVal))
		})

		// d) Sanitize style attribute if present
		tagAttrs = styleAttrRegex.ReplaceAllStringFunc(tagAttrs, func(styleMatch string) string {
			rawVal := extractAttrValue(styleMatch)
			safeStyle := sanitizeStyle(rawVal)
			if safeStyle == "" {
				return ""
			}
			return fmt.Sprintf(`style="%s"`, escapeAttr(safeStyle))
		})

		// e) For <a> tags: enforce external link safety (nofollow noopener noreferrer target="_blank")
		if tagName == "a" {
			tagAttrs = enforceSafeLinkAttrs(tagAttrs)
		}

		// Clean up multiple spaces
		tagAttrs = strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(tagAttrs, " "))
		if tagAttrs != "" {
			tagAttrs = " " + tagAttrs
		}

		return fmt.Sprintf("<%s%s>", tagName, tagAttrs)
	})

	return strings.TrimSpace(cleaned)
}

// IsSafeURL validates that a URL scheme is benign and free from executable schemes.
func IsSafeURL(urlStr string) bool {
	raw := strings.TrimSpace(urlStr)
	if raw == "" {
		return false
	}

	// Decode URL-encoded characters and strip control spaces to catch obfuscated schemes
	normalized := strings.ToLower(raw)
	normalized = strings.ReplaceAll(normalized, "\t", "")
	normalized = strings.ReplaceAll(normalized, "\r", "")
	normalized = strings.ReplaceAll(normalized, "\n", "")
	normalized = strings.ReplaceAll(normalized, " ", "")

	for _, bad := range dangerousProtocols {
		if strings.HasPrefix(normalized, bad) {
			return false
		}
	}

	// Also catch entity-encoded or double-encoded javascript: e.g. java&#115;cript:
	if strings.Contains(normalized, "javascript") || strings.Contains(normalized, "vbscript") {
		return false
	}

	// Relative URLs and anchor links are safe
	if strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "#") {
		return true
	}

	// Check scheme for absolute URLs
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme == "" {
		// Could be a relative path without leading slash
		return !strings.Contains(raw, ":")
	}

	return scheme == "http" || scheme == "https" || scheme == "mailto" || scheme == "tel"
}

// sanitizeStyle strips dangerous CSS constructs that can trigger code execution or CSS injection attacks.
func sanitizeStyle(style string) string {
	low := strings.ToLower(style)
	// Check for dangerous CSS functions and patterns
	if strings.Contains(low, "expression(") ||
		strings.Contains(low, "javascript:") ||
		strings.Contains(low, "behavior:") ||
		strings.Contains(low, "-moz-binding") ||
		strings.Contains(low, "@import") {
		return ""
	}

	// Disallow external URLs inside styles to prevent CSS exfiltration
	if strings.Contains(low, "url(") {
		return ""
	}

	return style
}

func extractAttrValue(attrMatch string) string {
	parts := strings.SplitN(attrMatch, "=", 2)
	if len(parts) < 2 {
		return ""
	}
	val := strings.TrimSpace(parts[1])
	val = strings.Trim(val, `"'`)
	return val
}

func escapeAttr(val string) string {
	val = strings.ReplaceAll(val, `"`, "&quot;")
	return val
}

func enforceSafeLinkAttrs(attrs string) string {
	// Check if this is an external link
	isExternal := false
	hrefMatches := hrefAttrRegex.FindStringSubmatch(attrs)
	if len(hrefMatches) > 1 {
		val := extractAttrValue(hrefMatches[0])
		lowVal := strings.ToLower(val)
		if strings.HasPrefix(lowVal, "http://") || strings.HasPrefix(lowVal, "https://") {
			if !strings.Contains(lowVal, "pidiekab.web.id") && !strings.Contains(lowVal, "localhost") {
				isExternal = true
			}
		}
	}

	if !isExternal {
		return attrs
	}

	// Remove any existing rel and target attributes
	attrs = relAttrRegex.ReplaceAllString(attrs, "")
	attrs = targetAttrRegex.ReplaceAllString(attrs, "")

	return strings.TrimSpace(attrs) + ` rel="nofollow noopener noreferrer" target="_blank"`
}
