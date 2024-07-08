# Go-Link-Shortener
---
This project is a simple bitly clone that shortens URLs. The backend is written in go and the frontend in HTMX.

This is still a WIP, below is my current TODO for the project 

Main Features/Issues to work on:
* Add captcha to the url shortener form
* A function to validate a signed in user on a non-protected route so that signed in users don't get shown a captcha for the url shortener 
* Make it so that expired sessions get removed from the databse at some point
* Make navbar interactive to show if you are logged in, only display register button if not logged in
* Link statistics for signed in users (Will need to figure out how to store on the backend)
    - Total clicks
    - 7, 30, 90, 365 day totals
    - A list of every request made with origin IP, User agent, and other applicable info
    - Exactly what can be seen should be configureable by server admin (for privacy)
* QR code images for links


Small fixes:
* standardize order of handler func parameters
* Modify UserLogin in cmd/ to use the pattern used in internal/sessmngt/database_funcs for UserSession
* Ability for admin to not allow creation of links unless user is signed in
* Make logging work to a file, and errors be more descriptive for where they happened
* Add option for auto TLS, needs work on config for all fields
* Redirect http to https
* Check all handler functions to change any passing in of db, to using the db stored in echo.Context.
    - How I do it now is fine, but I had to make a middleware to get the database to the authentication middleware so I might as well take advantage 
* Make the shortcode forms allow the submission of both links with https:// to start and without, not sure if this is nessicary 
* Store the strings that get used as keys in sess.Values[] in a const somewhere for consistency


## Installation
---
1. Run make
2. Run the binary generated
