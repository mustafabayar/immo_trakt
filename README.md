# ImmoTrakt
Just a hobby flat tracker for **ImmoScout24**. Finds offers according to your search and sends message via Telegram Bot.
There is no web scraping, it works via API calls.

## How to setup config?
- - -
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

### **telegram.chat_id**:
To get your chat id, you need to send a message to the newly registered bot. After that you can use the following curl command to get the chat id:
```
$ curl https://api.telegram.org/bot[TELEGRAM_BOT_TOKEN]/getUpdates
```
The result will look like this:
```
{"ok":true,"result":[{"update_id":123123123,
"message":{"message_id":123,"from":{"id":XXXXXXXX,"is_bot":false,"first_name":"YOUR_NAME","language_code":"en"},"chat":{"id":111111111,"first_name":"YOUR_NAME","type":"private"},"date":1231231231,"text":"XYZ"}}]}
```
You need to copy the id from chat varieble, not message. So it is **111111111** in the above case.

## How to run?
- - -
```
go run main.go
```
or 

on Windows run the executable **immo-trakt.exe**

on Macos run the **immo-trakt** binary (Make sure it is executable by running "chmod +x path-to-binary")