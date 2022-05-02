# wakeru

Wakeru is a simplistic, small-scale file upload and sharing service written in Go. It is meant to be self hosted and used mainly for family & friends.

Authentication is done via generated tokens instead of classic username + password accounts. This approach was inspired by [https://mullvad.net/en/](Mullvad)'s generated user tokens. To ensure not giving out the full token randomly, only a part of the uploader's token is shown when serving files.

This service does not require a database as it only uses the local filesystem.

---

As this project is still very much a work in progress, any sort of documentation is lacking. This will be tackled after approaching the completion of initially planned features.
