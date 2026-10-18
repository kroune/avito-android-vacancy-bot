# avito-android-vacancy-bot

Бот, который раз в час проверяет вакансии на
[career.avito.com](https://career.avito.com/vacancies/?q=&action=filter&direction=razrabotka)
и присылает уведомление в Telegram, когда появляется **новая** вакансия,
связанная с Android.

## Как работает

1. Скачивает HTML страницы вакансий (сайт отдаёт server-side HTML, JS не нужен).
2. Парсит список вакансий и фильтрует по ключевым словам (по умолчанию `android`,
   без учёта регистра — ищется в названии и команде).
3. Сравнивает с уже увиденными (файл состояния `STATE_FILE`) и шлёт в Telegram
   только новые, через Telegram Bot API.

## Конфигурация (env)

| Переменная           | По умолчанию                                             | Описание                          |
|----------------------|----------------------------------------------------------|-----------------------------------|
| `TELEGRAM_BOT_TOKEN` | — (обязательно)                                          | Токен бота от @BotFather          |
| `TELEGRAM_CHAT_ID`   | — (обязательно)                                          | ID чата (@userinfobot)            |
| `PAGE_URL`           | `https://career.avito.com/vacancies/?...razrabotka`      | Страница вакансий                 |
| `CHECK_INTERVAL`     | `1h`                                                     | Интервал проверки                 |
| `KEYWORDS`           | `android`                                                | Ключевые слова через запятую      |
| `STATE_FILE`         | `/data/state.json`                                       | Файл состояния (seen IDs)         |
| `NOTIFY_ERRORS`      | `false`                                                  | Слать в ТГ ошибки проверок        |

## Локальный запуск

```bash
export TELEGRAM_BOT_TOKEN=... TELEGRAM_CHAT_ID=...
STATE_FILE=./state.json go run ./cmd/bot
```

## Docker

```bash
docker build -t avito-vacancy-bot .
docker run -e TELEGRAM_BOT_TOKEN=... -e TELEGRAM_CHAT_ID=... -v $(pwd)/data:/data avito-vacancy-bot
```

## CI / GitHub Packages

При пуше в `main` GitHub Actions собирает образ и публикует его как package:
`ghcr.io/kroune/avito-android-vacancy-bot:latest` (и тег `sha-<commit>`).
Образ публично не виден, пока не сделать package public в настройках GitHub —
для приватного образа в кубере нужен `imagePullSecret` с PAT (scope `read:packages`).

## Деплой в Kubernetes

```bash
cd deploy
cp secret.example.yaml secret.yaml  # заполнить токен и chat_id
kubectl apply -f secret.yaml
kubectl apply -k .
```

Состояние хранится на PVC (`deploy/pvc.yaml`), поэтому после рестарта пода
бот не будет повторно слать старые вакансии.
