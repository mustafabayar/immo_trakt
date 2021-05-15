<p align="center">
  <img src="https://imgur.com/C1aP2p5.png">
</p>

<p align="left">
<img src="https://img.shields.io/github/license/mustafabayar/immo_trakt">
<img src="https://img.shields.io/maintenance/yes/2021">
<a href="https://www.codacy.com/gh/mustafabayar/immo_trakt/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=mustafabayar/immo_trakt&amp;utm_campaign=Badge_Grade"><img src="https://app.codacy.com/project/badge/Grade/1b1ae1e6c305418d91da7c9c4c7d9adf"/></a>
<a href="http://golang.org"><img src="https://img.shields.io/badge/Made%20with-Go-1f425f.svg"/></a>
</p>

Just a hobby flat tracker for <span style="color:#01ffd1">**ImmoScout24**</span>. Finds offers according to your search and sends message via Telegram Bot.
There is no web scraping, it works via API calls.

## [How to setup config](#how-to-setup-config)
Copy the `config_skeleton.yml` and save it as `config.yml`. Then fill all the config parameters inside the config file.

-   `immo_trakt.frequency:`
Duration string as described in https://golang.org/pkg/time/#ParseDuration. For example ***1m*** means every 1 minute.

-   `immo_trakt.include_existing_offers:`
    -   `true` if you want the bot to send message for all the existing offers that fits the given search url. This will also send the new offers that are being added after app started, but you will also receive message for every single existing offer for the given search url.
    -   `false` if you want the bot to only start sending offers that are added after the app start running. This mode won't send the offers that were already exist before the app started. **false** makes more sense to not clutter your chat with tons of already existing offers that fits your search url. Because for most people, use-case of the bot is to see the ***new*** offers as soon as possible, not the existing ones as you can already see them when you open the immobilien scout.

-   `immobilien_scout.search:`
Simply go to immobilien scout and make a search according to your criterias and then copy the final url to this config value.

-   `immobilien_scout.exclude_wbs:`
    -   `true` if you want offers that contains **WBS** keyword in title to be ignored. 
    -   `false` otherwise.

-   `immobilien_scout.exclude_tausch:`
    -   `true` if you want offers that contains **TAUSCH** keyword in title to be ignored. 
    -   `false` otherwise.

-   `telegram.token:`
Register a new bot with the [BotFather](https://telegram.me/BotFather). Follow the instructions and create your bot. 
Botfather will return bot token to access the HTTP API.

Before running the application make sure to first send a message to the created bot on Telegram so that the application can detect which chat to send messages.

## [How to run](#how-to-run)

### Locally with Go
Install [Go](https://golang.org/doc/install) if you don't already have it.
```
go run main.go
```
- - -
### Locally with Docker
Make sure to follow [How to setup config?](#how-to-setup-config) section and copy the `config.yml` file first. Docker is going to use that file. Afterwards:
```
docker build -t immo-trakt .
docker run -d immo-trakt
```
Don't forget to check the logs of the running docker container. If you see this error:

`Telegram chat not found, please send a message to the bot first and try to run the ImmoTrakt again!`

Just do what the error says and run it again.
- - -
### Deploy to cloud
Make sure to follow [How to setup config?](#how-to-setup-config) section and copy the `config.yml` file. Then deploy to any of the cloud services you prefer. Here I will describe how to use [Heroku](https://www.heroku.com/pricing) to run this app free for 7/24:
1.  Install [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli) and [Docker](https://docs.docker.com/get-docker/)
2.  Login to heroku: 
```
heroku login
```
3.  Log in to Container Registry: 
```
heroku container:login
``` 
4.  Create app on Heroku: 
```
heroku create
```
> This will give output like: ```Creating cool-app-name... done, stack is heroku-18```
5.  Build the image and push to Container Registry(Use the app name created from previous step): 
```
heroku container:push worker --app cool-app-name
```
6.  Then release the image to your app: 
```
heroku container:release worker --app cool-app-name
```
This will build the docker image from Dockerfile and release the app on Heroku. Go to your [dashboard](https://dashboard.heroku.com/apps) to view the app. After entering the app dashboard you can see the logs of running app by clicking ```More``` -> ```View logs``` on top right corner. Make sure in the Dyno formation widget, worker dyno is ```ON``` and there is no web dyno running.

Normally Heroku free tier sleeps after 30 minutes of inactivity (Not receiving any web traffic). But this is only valid for web dynos, our app doesn't serve any endpoints therefore we don't need a web dyno, all we need is a background app. Worker dyno is best for this use-case as it doesn't go to sleep. Normally Heroku free tier gives 550 hours of free dyno usage when you register, which means you can not run it 7/24 for whole month. But if you add your credit card information they will increase the free dyno limit to 1000 hours which is more than enough to run an app for entire month.