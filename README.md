# Go-Link-Shortener
---
This project is a simple bitly clone that shortens URLs. The backend is written in go and the frontend in HTMX.

This is still a WIP, below is the features that will be implemented in the future:
* User account creation/authentication
* Ability for admin to not allow creation of links unless user is signed in
* Link statistics for signed in users (Will need to figure out how to store on the backend)
* QR code images for links

* Add option for auto TLS, needs work on config for all fields
* Redirect http to https

* Make logging work to a file, and errors be more descriptive for where they happened

* standardize order of handler func parameters
## Installation
---
1. Run make
2. Run the binary generated
