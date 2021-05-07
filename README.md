# ImmoTrakt
Just a hobby flat tracker for **ImmoScout24**. Finds offers according to your search and sends message via Telegram Bot.
There is no web scraping, it works via API calls.

## [How to setup config?](#how-to-setup-config)
Copy the **config_skeleton.yml** and save it as **config.yml**. Then fill all the config parameters inside the config file.

### **immo_trakt.frequency**:
Duration string as described in https://golang.org/pkg/time/#ParseDuration. For example 1m means every 1 minute.

### **immo_trakt.include_existing_offers**:
**True** if you want the bot to send message for all the existing offers that fits the given criteria.
**False** if you want the bot to only start sending offers that are added after the app start running.

### **immobilien_scout.search**: 
Simply go to immobilien scout and make a search according to your criterias and then copy the final url to this config value.

### **immobilien_scout.exclude_wbs**: 
**True** if you want offers that contains **WBS** keyword in title to be ignored. **False** otherwise.

### **immobilien_scout.exclude_tausch**: 
**True** if you want offers that contains **TAUSCH** keyword in title to be ignored. **False** otherwise.

### **telegram.token**:
Register a new bot with the [BotFather](https://telegram.me/BotFather). Follow the instructions and create your bot. 
Botfather will return bot token to access the HTTP API.

Before running the application make sure to first send a message to the created bot on Telegram so that the application can detect which chat to send messages.

## [How to run?](#how-to-run)
### With Docker:
Make sure to follow [How to setup config?](#how-to-setup-config) section and copy the `config.yml` file first. Docker is going to use that file. Afterwards:
```
docker build -t immo-trakt .
docker run -d immo-trakt
```
Don't forget to check the logs of the running docker container. If you see this error:

`Telegram chat not found, please send a message to the bot first and try to run the ImmoTrakt again!`

Just do what the error says and run it again.

- - -

### Locally:
```
go run main.go
```