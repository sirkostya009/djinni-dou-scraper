# Job Scraper

Job scraper is a Telegram bot that sends you job offers from various sources.
It allows users to add up to 2 subscriptions, i.e. individual URLs to either a [Djinni](https://djinni.co) or
[Dou](https://jobs.dou.ua) job boards. The bot will then send you new job offers from these sources.

This is a service for handling user interaction.

There's also the scraper part, which is a separate service on its own:
[job-scraper-lambda](https://github.com/sirkostya009/job-scraper-lambda).

### How to run:
1. Make sure you have `TELEGRAM_BOT_TOKEN`, `PORT`, `WEBHOOK_URL` and `DATABASE_URL` environment variables set. With
`DATABASE_URL` defaulting to `postgres://postgres:postgres@localhost:5432/templ_htmx_go` if not set.
2. `WEBHOOK_URL` is the domain where the bot is hosted. If it's localhost you're running at make sure you got them
certificates installed, they can be self-singed.
3. After that just run `go run .`
