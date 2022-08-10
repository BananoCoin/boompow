[![License](https://img.shields.io/github/license/BananoCoin/boompow-next)](https://github.com/bananocoin/boompow/blob/master/LICENSE) [![CI](https://github.com/bananocoin/boompow/workflows/CI/badge.svg)](https://github.com/bananocoin/boompow/actions?query=workflow%3ACI)

<p align="center">
  <img src="https://raw.githubusercontent.com/BananoCoin/boompow-next/master/logo.svg" width="300">
</p>

This is a distributed proof of work system for the [BANANO](https://banano.cc) and [NANO](https://nano.org) cryptocurrencies.

## What is It?

Banano transactions require a "proof of work" in order to be broadcasted and confirmed on the network. Basically you need to compute a series of random hashes until you find one that is "valid" (satisifies the difficulty equation). This serves as a replacement for a transaction fee.

## Why do I want BoomPow?

The proof of work required for a BANANO transasction can be calculated within a couple seconds on most modern computers. Which begs the question "why does it matter?"

1. There's applications that require large volumes of PoW, while an individual calculation can be acceptably fast - it is different when it's overloaded with hundreds of problems to solve all at the same time.
   - The [Graham TipBot](https://github.com/bbedward/Graham_Nano_Tip_Bot) has been among the biggest block producers on the NANO and BANANO networks for more than a year. Requiring tens of thousands of calculations every month.
   - The [Twitter and Telegram TipBots](https://github.com/mitche50/NanoTipBot) also calculate PoW for every transaction
   - [Kalium](https://kalium.banano.cc) and [Natrium](https://natrium.io) are the most widely used wallets on the NANO and BANANO networks with more than 10,000 users each. They all demand PoW whenever they make or send a transaction.
   - There's many other popular casinos, exchanges, and other applications that can benefit from a highly-available, highly-reliable PoW service.
2. While a single PoW (for BANANO) can be calculated fairly quickly on modern hardware, there are some scenarios in which sub-second PoW is highly desired.
   - [Kalium](https://kalium.banano.cc) and [Natrium](https://natrium.io) are the top wallets for BANANO and NANO. People use these wallets to showcase BANANO or NANO to their friends, to send money when they need to, they're used in promotional videos on YouTube, Twitter, and other platforms. _Fast_ PoW is an absolute must for these services - the BoomPow system will provide incredibly fast proof of work from people who contribute using high-end hardware.

All of the aforementioned services will use the BoomPow system, and others services are free to request access as well.

## Who is Paying for this "High-End" Hardware?

[BANANO](https://banano.cc) is an instant, feeless, rich in potassium cryptocurrency. It has had an ongoing **free and fair** distribution since April 1st, 2018.

BANANO is distributed through [folding@home "mining"](https://bananominer.com), faucet games, giveaways, rain parties on telegram and discord, and more. We are always looking for new ways to distribute BANANO _fairly._

BoomPow is going to reward contributors with BANANO. Similar to mining, if you provide valid PoW solutions for the BoomPow system you will get regular payments based on how much you contribute.

## Components

This is a GOLang "monorepo" that contains all BoomPoW Services

- [Server](https://github.com/bananocoin/boompow/blob/master/apps/server)
- [Client](https://github.com/bananocoin/boompow/blob/master/apps/client)
- [Moneybags (Payment Cron)](https://github.com/bananocoin/boompow/blob/master/services/moneybags)

## Contributing

Development for BoomPoW is ideal with a [docker](https://www.docker.com/) development environment.

These alias may be helpful to add to your `~/.bashrc` or `~/.zshrc`

```
alias dcup="docker-compose up -d"
alias dcdown="docker-compose down"
alias dcbuild="docker-compose build"
alias dczsh="docker-compose exec app /bin/zsh"
alias dcps="docker-compose ps"
alias dcgo="docker-compose exec app go"
alias dcgoclient="docker-compose exec --workdir /app/apps/client app go"
alias dcgoserver="docker-compose exec --workdir /app/apps/server app go"
alias psql-boompow="PGPASSWORD=postgres psql boompow -h localhost -U postgres -p 5433"
```

Once you have docker installed and running, as well as docker-compose, you can start developing with:

```
> dcup
# To run the server
> dcgo run github.com/bananocoin/boomow-next/apps/server server
# To run the client
> dcgo run github.com/bananocoin/boompow/apps/client
```

To get an interactive shell in the container

```
dczsh
```

## Issues

For issues, create tickets on the [Issues Page](https://github.com/bananocoin/boompow/issues)

The [BANANO discord server](https://chat.banano.cc) has a channel dedicated to to boompow if you have general questions about the service.
