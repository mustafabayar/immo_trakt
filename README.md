<p align="center">
  <img src="https://imgur.com/C1aP2p5.png">
</p>

<p align="left">
<img src="https://img.shields.io/github/license/mustafabayar/immo_trakt">
<img src="https://img.shields.io/maintenance/yes/2022">
<a href="https://www.codacy.com/gh/mustafabayar/immo_trakt/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=mustafabayar/immo_trakt&amp;utm_campaign=Badge_Grade"><img src="https://app.codacy.com/project/badge/Grade/1b1ae1e6c305418d91da7c9c4c7d9adf"/></a>
<a href="http://golang.org"><img src="https://img.shields.io/badge/Made%20with-Go-1f425f.svg"/></a>
</p>

A flat tracker for <span style="color:#01ffd1">**ImmoScout24**</span>. Finds offers according to your search and sends message via Telegram Bot.
There is no web scraping, it works via API calls.

---
##### Table of Contents  
-   [How to setup config](#how-to-setup-config)
-   [How to run](#how-to-run)
    -   [Run locally with Go](#run-locally-with-go)
    -   [Run locally with Docker](#run-locally-with-docker)
    -   [Deploy to cloud](#deploy-to-cloud)
-   [Troubleshooting](#troubleshooting)
-   [License](#license)
---
## [How to setup config](#how-to-setup-config)
You can configure ImmoTrakt in two ways:
1.  Copy the `config_skeleton.yml` and save it as `config.yml`. Then fill all the config parameters inside the config file.
2.  Set the config parameters as environment variables. See the necessary keys below.

Using combination of two is also okay. For example, you can set some values in config file and some others in ENV. But ENV variables will always have priority over config file if they are set.

| Config Key | ENV key | Description | Default Value |
| ------------------------------------ | ---------------------------- | ----------- | ------------- |
| `immo_trakt.frequency` | `IMMOTRAKT_FREQUENCY` | Duration string as described in https://golang.org/pkg/time/#ParseDuration. For example ***1m*** means every 1 minute. | **1m** |
| `immo_trakt.include_existing_offers` | `IMMOTRAKT_INCLUDE_EXISTING` | `true` if you want the bot to send message for all the existing offers that fits the given search url. `false` if you want the bot to only start sending offers that are added after the app started running. For most people, use-case of this bot is to see the ***new*** offers as soon as possible, not the existing ones as you can already see them when you open the ImmoblienScout website. Therefore **false** makes more sense to not clutter your chat with tons of already existing offers. | **false** |
| `immobilien_scout.search` | `IMMOTRAKT_SEARCH` | Simply go to immobilien scout and make a search according to your criterias and then copy the final url to this config value. | None. Must be set. |
| `immobilien_scout.exclude_wbs` | `IMMOTRAKT_EXCLUDE_WBS` | `true` if you want offers that contains **WBS** keyword in the offer title to be ignored. `false` otherwise. | **false** |
| `immobilien_scout.exclude_tausch` | `IMMOTRAKT_EXCLUDE_TAUSCH` | `true` if you want offers that contains **TAUSCH** keyword in the offer title to be ignored. `false` otherwise. | **false** |
| `immobilien_scout.exclude_senior` | `IMMOTRAKT_EXCLUDE_SENIOR` | `true` if you want offers that contains **SENIOR** keyword in the offer title to be ignored. `false` otherwise. | **false** |
| `telegram.token` | `IMMOTRAKT_TELEGRAM_TOKEN` | Register a new bot with the [BotFather](https://telegram.me/BotFather). Follow the instructions and create your bot. Botfather will return bot token to access the HTTP API. | None. Must be set. |
| `telegram.chat_id` | `IMMOTRAKT_TELEGRAM_CHAT_ID` | This is the personal chat you have with the created Bot. It is okay to not know your chat ID, just leave it empty and we will try to find the correct chat ID using the token. | None. Leave it empty if you don't know. |

:exclamation: Before running the application make sure to first send a message to the created bot on Telegram so that the application can detect which chat to send messages.
When you run the application and we retrieved the chat ID, it will be logged as:
> Telegram chat ID found as XXXXXXXXX

At that point you can take this chat ID and set it to `IMMOTRAKT_TELEGRAM_CHAT_ID` environment variable. This would make it very convenient because when you restart the application after some time, then you would have to do the same process of sending message to Telegram so that chat ID can be retrieved from API. But if you store the chat ID in config or env variable, you won't need to touch it again.

## [How to run](#how-to-run)

### [Run locally with Go](#run-locally-with-go)
Make sure to follow [How to setup config](#how-to-setup-config) section to set configuration values properly.

Install [Go](https://golang.org/doc/install) if you don't already have it. Afterwards execute the following command in project root folder:
```
go run main.go
```
---
### [Run locally with Docker](#run-locally-with-docker)
Make sure to follow [How to setup config](#how-to-setup-config) section to set configuration values properly.

Afterwards execute the following commands on project root folder:
```
docker build -t immo-trakt .
docker run -d immo-trakt
```
Don't forget to check the logs of the running docker container.

---
### [Deploy to cloud](#deploy-to-cloud)
Make sure to follow [How to setup config](#how-to-setup-config) section to set configuration values properly.

You can deploy the app to any of the cloud services you prefer. Here I will describe how to use [Heroku](https://www.heroku.com/pricing) to run this app free for 7/24:
1.  Install [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli) and [Docker](https://docs.docker.com/get-docker/)
2. Start the docker engine as it is needed for the following commands.
3.  Login to heroku: 
```
heroku login
```
4.  Log in to Container Registry: 
```
heroku container:login
``` 
5.  Create app on Heroku: 
```
heroku create
```
> This will give output like: `Creating cool-app-name... done, stack is heroku-18`
6.  Build the image and push to Container Registry(Use the app name created from previous step): 
```
heroku container:push worker --app cool-app-name
```
7.  Then release the image to your app: 
```
heroku container:release worker --app cool-app-name
```
This will build the docker image from Dockerfile and release the app on Heroku. Go to your [dashboard](https://dashboard.heroku.com/apps) to view the app. After entering the app dashboard you can see the logs of running app by clicking `More` -> `View logs` on top right corner.
In the logs, you should see your Telegram chat ID if the app was able to find it.

8. (Optional) Go to `Settings` -> `Config Vars` and add a new config var key as `IMMOTRAKT_TELEGRAM_CHAT_ID` and value as the found chat ID from the logs.

Also make sure in the Dyno formation widget, worker dyno is `ON` and there is no web dyno running.
Normally Heroku free tier sleeps after 30 minutes of inactivity (Not receiving any web traffic). But this is only valid for web dynos, our app doesn't serve any endpoints therefore we don't need a web dyno, all we need is a background app. Worker dyno is best for this use-case as it doesn't go to sleep. Normally Heroku free tier gives 550 hours of free dyno usage when you register, which means you can not run it 7/24 for whole month. But if you add your credit card information they will increase the free dyno limit to 1000 hours which is more than enough to run an app for entire month.

## [Troubleshooting](#troubleshooting)

Always check the logs of the running app, most of the error messages are quite straightforward.

If you see this error:

```
Telegram chat not found, please first send a message to the bot on Telegram and then try to run the ImmoTrakt again!
```

Just do what the message says and run it again.

If you see any problem regarding the code such as parsing error etc., please open an issue and I will work on it as soon as I can.

## [License](#license)
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
