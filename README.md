# Go-Link-Shortener
---
This project is a simple bitly clone that shortens URLs. The backend is written in go and the frontend in HTMX.

This is still a WIP, below is the features that will be implemented in the future:

Urgent:
* Either logout doesn't work or my page restriction is broken. Figure this out
* Consolidation of all user/session management to its own go file
* Redirection to user page after registration (Do after session management is consolidated)
* Server side validation of login sessions
* Make navbar interactive to show if you are logged in, only display register button if not logged in
* Link statistics for signed in users (Will need to figure out how to store on the backend)
* QR code images for links


QOL:
* Set database file name/path with config
* standardize order of handler func parameters
* hCaptcha on register page
* Make an error template on each page so that error messages can be streamlined
* Ability for admin to not allow creation of links unless user is signed in
* Make logging work to a file, and errors be more descriptive for where they happened
* Add option for auto TLS, needs work on config for all fields
* Redirect http to https



## Installation
---
1. Run make
2. Run the binary generated
