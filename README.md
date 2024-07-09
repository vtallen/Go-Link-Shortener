# Go-Link-Shortener
---
This project is a simple bitly clone that shortens URLs and counts how many times those URLS are accessed. 

## Features
---
* Allows the user to created shorted links of any url
* User accounts
    - Allows the tracking of the number of clicks on a URL
    - Allows the deletion of shortlinks created by a user
* hCaptcha on all forms to ensure the webapp is resistant to bot form submissions

## Technologies used
---
* Golang
    - Echo (web server)
* HTMX
* SQlite

## Setup and Usage 
---
This server has been tested and verified for use only on Linux. You may run into unexpected issues running it on anything else.

1. Make a copy of config_template.yaml and rename it to config.yaml
2. Generate tls certificates, then modify config.yaml to have the paths of your certificate and key files
3. Go to https://www.hcaptcha.com/
    - Create an account
    - Copy your secret key and place it in config.yaml
    - Create a site key and place it in config.yaml
4. Generate a strong, random password to use as the cookie secret. Place this in config.yaml
5. Run make to generate an executable
6. Run the server ```./server```

## TODO
---
* Document all functions
* Replace any print statements with log statements
* Add logging to the handlers where needed
* Enable auto tls
