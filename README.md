# Job Scraper

Job scraper is a Telegram bot that sends you job offers from various sources.
It allows users to add up to 2 subscriptions, i.e. individual URLs to either a [Djinni](https://djinni.co) or
[Dou](https://jobs.dou.ua) job boards. The bot will then send you new job offers from these sources.

This is a service for handling user interaction.

There's also the scraper part, which is a separate service on its own:
[job-scraper-lambda](https://github.com/sirkostya009/job-scraper-lambda).

### How to run:
1. Make sure you have `TELEGRAM_BOT_TOKEN` environment variable set. And `MONGO_URL`, but if you don't, it will default
to `mongodb://localhost:27017`.
2. After that just run `go run .`
