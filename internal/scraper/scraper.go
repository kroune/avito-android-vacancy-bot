// Package scraper fetches and parses vacancies from career.avito.com.
package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Vacancy is a single parsed job posting from the vacancies page.
type Vacancy struct {
	ID     string // stable identifier (data-vacancy-id)
	Name   string
	URL    string // absolute link to the vacancy page
	City   string
	Team   string
	Remote string
}

// Matches reports whether the vacancy name or team contains any of the
// given keywords (case-insensitive).
func (v Vacancy) Matches(keywords []string) bool {
	haystack := strings.ToLower(v.Name + " " + v.Team)
	for _, kw := range keywords {
		if strings.Contains(haystack, strings.ToLower(strings.TrimSpace(kw))) {
			return true
		}
	}
	return false
}

const userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36"

// Fetch downloads pageURL and parses all vacancies from it.
func Fetch(ctx context.Context, client *http.Client, pageURL string) ([]Vacancy, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch page: unexpected status %s", resp.Status)
	}

	return Parse(resp.Body, pageURL)
}

// Parse extracts vacancies from the HTML page served at pageURL
// (pageURL is used as the base for resolving relative vacancy links).
func Parse(r io.Reader, pageURL string) ([]Vacancy, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("parse page: %w", err)
	}

	var out []Vacancy
	doc.Find(".vacancies-section__item").Each(func(_ int, s *goquery.Selection) {
		link := s.Find("a.vacancies-section__item-name").First()
		name := strings.TrimSpace(link.Text())
		href, _ := link.Attr("href")
		if name == "" || href == "" {
			return
		}
		id := s.AttrOr("data-vacancy-id", href)
		out = append(out, Vacancy{
			ID:     id,
			Name:   name,
			URL:    absoluteURL(pageURL, href),
			City:   s.AttrOr("data-vacancy-geo", strings.TrimSpace(s.Find(".vacancies-section__item-cities").Text())),
			Team:   s.AttrOr("data-vacancy-team", ""),
			Remote: s.AttrOr("data-vacancy-remote", ""),
		})
	})
	return out, nil
}

func absoluteURL(pageURL, href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	// scheme + host from pageURL
	if i := strings.Index(pageURL, "://"); i >= 0 {
		rest := pageURL[i+3:]
		if j := strings.IndexByte(rest, '/'); j >= 0 {
			return pageURL[:i+3] + rest[:j] + href
		}
		return pageURL + href
	}
	return href
}
