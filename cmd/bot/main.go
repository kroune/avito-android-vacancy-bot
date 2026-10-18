// Command bot polls career.avito.com for Android vacancies and notifies a Telegram chat.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kroune/avito-android-vacancy-bot/internal/notifier"
	"github.com/kroune/avito-android-vacancy-bot/internal/scraper"
	"github.com/kroune/avito-android-vacancy-bot/internal/state"
)

type config struct {
	pageURL      string
	interval     time.Duration
	stateFile    string
	keywords     []string
	botToken     string
	chatID       string
	notifyErrors bool
}

func loadConfig() config {
	return config{
		pageURL:      env("PAGE_URL", "https://career.avito.com/vacancies/?q=&action=filter&direction=razrabotka"),
		interval:     envDuration("CHECK_INTERVAL", time.Hour),
		stateFile:    env("STATE_FILE", "/data/state.json"),
		keywords:     strings.Split(env("KEYWORDS", "android"), ","),
		botToken:     os.Getenv("TELEGRAM_BOT_TOKEN"),
		chatID:       os.Getenv("TELEGRAM_CHAT_ID"),
		notifyErrors: env("NOTIFY_ERRORS", "false") == "true",
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		log.Printf("invalid %s %q, using default %s", key, v, def)
	}
	return def
}

func main() {
	cfg := loadConfig()
	if cfg.botToken == "" || cfg.chatID == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID are required")
	}

	store, err := state.Load(cfg.stateFile)
	if err != nil {
		log.Fatalf("load state: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	tg := &notifier.Telegram{Token: cfg.botToken, ChatID: cfg.chatID, Client: client}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("started: url=%s interval=%s keywords=%v", cfg.pageURL, cfg.interval, cfg.keywords)

	check := func() {
		if err := checkOnce(ctx, cfg, client, store, tg); err != nil {
			log.Printf("check failed: %v", err)
			if cfg.notifyErrors {
				if nerr := tg.SendText(ctx, "⚠️ avito-vacancy-bot: "+err.Error()); nerr != nil {
					log.Printf("send error notification: %v", nerr)
				}
			}
		}
	}

	check() // run immediately on startup
	ticker := time.NewTicker(cfg.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Print("shutting down")
			return
		case <-ticker.C:
			check()
		}
	}
}

// checkOnce fetches the page once and notifies about new matching vacancies.
func checkOnce(ctx context.Context, cfg config, client *http.Client, store *state.Store, tg *notifier.Telegram) error {
	vacancies, err := scraper.Fetch(ctx, client, cfg.pageURL)
	if err != nil {
		return err
	}
	log.Printf("fetched %d vacancies", len(vacancies))

	sent := 0
	for _, v := range vacancies {
		if !v.Matches(cfg.keywords) {
			continue
		}
		if !store.IsNew(v.ID) {
			continue
		}
		log.Printf("new matching vacancy: %s (%s)", v.Name, v.URL)
		if err := tg.SendVacancy(ctx, v); err != nil {
			return err // do not save state; retry on next tick
		}
		sent++
	}
	if err := store.Save(); err != nil {
		return err
	}
	log.Printf("check done: %d new notifications sent", sent)
	return nil
}
