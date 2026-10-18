// Package notifier sends notifications about new vacancies to Telegram.
package notifier

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kroune/avito-android-vacancy-bot/internal/scraper"
)

// Telegram sends messages through a Telegram bot.
type Telegram struct {
	Token  string
	ChatID string
	Client *http.Client
}

// SendVacancy formats and sends a message about a single vacancy.
func (t *Telegram) SendVacancy(ctx context.Context, v scraper.Vacancy) error {
	var b strings.Builder
	b.WriteString("\U0001F4E2 Новая Android-вакансия в Авито!\n\n")
	fmt.Fprintf(&b, "<b>%s</b>\n", htmlEscape(v.Name))
	if v.Team != "" {
		fmt.Fprintf(&b, "Команда: %s\n", htmlEscape(v.Team))
	}
	if v.City != "" {
		fmt.Fprintf(&b, "Локация: %s\n", htmlEscape(v.City))
	}
	if v.Remote != "" {
		fmt.Fprintf(&b, "Удалённо: %s\n", htmlEscape(v.Remote))
	}
	fmt.Fprintf(&b, "\n<a href=\"%s\">Открыть вакансию</a>", htmlEscape(v.URL))
	return t.send(ctx, b.String())
}

// SendText sends a plain text message (used for startup/error notifications).
func (t *Telegram) SendText(ctx context.Context, text string) error {
	return t.send(ctx, htmlEscape(text))
}

func (t *Telegram) send(ctx context.Context, htmlText string) error {
	client := t.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	form := url.Values{
		"chat_id":                  {t.ChatID},
		"text":                     {htmlText},
		"parse_mode":               {"HTML"},
		"disable_web_page_preview": {"true"},
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.Token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram send: unexpected status %s", resp.Status)
	}
	return nil
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}
