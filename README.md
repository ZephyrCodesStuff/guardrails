# `guardrails`

It's a DNS server that allows you to block certain domains based on time of day and day of week.

Simple as that. No bells and whistles, no fancy UI, just a config file and a TCP & UDP server. (The fanciest thing you'll find is the logger.)

### ❓ Why?

If you ever found yourself trying to focus but keep getting distracted by social media, or you simply want to escape FOMO knowing that there are no notifications unread (they're literally blocked by the DNS server), this is for you.

I made this because time trackers aren't enough: it's far too simple to just press "15 more minutes please" and continue scrolling.

SSHing into my server and reconfiguring this, or changing the DNS settings just for a quick scroll, is a much higher barrier to entry, enough to make you rethink about scrolling or checking your notifications.

### 🛠️ How to set up

1. Get [Docker](https://docs.docker.com/get-docker/) (along with Docker Compose) installed.
2. (Optional) Clone the repository -- you can also just download the `docker-compose.yml` and `example.toml` files.
3. Copy `example.toml` into `config.toml` and edit it to your liking.
4. Run `docker-compose up -d` to start the server.

You may host this on your Raspberry Pi, a VPS of your own or anywhere you'd like. As long as your device can reach this, it'll do.

### 🧑‍💻 Authors

Me ([@zeph](https://github.com/ZephyrCodesStuff)).

Nope, not even Claude or ChatGPT helped me write this. Where's the fun in that? (Plus I wanted to learn Go.)
