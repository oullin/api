package markdown

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/oullin/pkg/portal"
	"gopkg.in/yaml.v3"
)

func (p Parser) Fetch() (string, error) {
	req, err := http.NewRequest("GET", p.Url, nil)

	if err != nil {
		return "", err
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch markdown: status %d", resp.StatusCode)
	}

	body, err := portal.ReadWithSizeLimit(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Parse splits the document into front-matter and content, then parses YAML.
// It also extracts a leading Parser image (header image) if present.
func Parse(data string) (*Post, error) {
	var post Post

	// Expecting format: ---\n<yaml>---\n<content>
	sections := strings.SplitN(data, "---", 3)
	if len(sections) < 3 {
		return nil, fmt.Errorf("invalid front-matter format")
	}

	fm := strings.TrimSpace(sections[1])
	body := strings.TrimSpace(sections[2])

	// Unmarshal YAML into FrontMatter
	err := yaml.Unmarshal([]byte(fm), &post.FrontMatter)
	if err != nil {
		return nil, err
	}

	// Look for a header image at the top of the content
	// Parser image syntax: ![alt](url)
	re := regexp.MustCompile(`^!\[(.*?)\]\((.*?)\)`)

	// Split first line from rest of content
	parts := strings.SplitN(body, "\n", 2)
	first := strings.TrimSpace(parts[0])

	if m := re.FindStringSubmatch(first); len(m) == 3 {
		post.ImageAlt = m[1]
		post.ImageURL = m[2]

		// Remaining content excludes the header image line
		if len(parts) > 1 {
			post.Content = strings.TrimSpace(parts[1])
		} else {
			post.Content = ""
		}
	} else {
		// No header image found; the entire body is content
		post.ImageAlt = ""
		post.ImageURL = ""
		post.Content = body
	}

	return &post, nil
}
