# External itch.io GameJam Voting System
This project is a standalone go program that can be used to manually manage voting for itch.io with no weighting algorithms, just averages.

## Setup
Create a `config.json` file with at least the following:

```json
{
  "Address": ":3000",
  "GameJam": "my-game-jam",
  "ClientID": "<acquired from itch.io OAuth setup>",
  "OAuthRedirect": "<OAuth redirect url provided in OAuth setup>",
  "VotingEnabled": true,
  "VotingFinished": false
}
```

**Address** is the address + port to listen for HTTP connections on.

**GameJam** is the game jam's name, as derived from the jam's URL. e.g, "https://itch.io/jam/ebiten-game-jam" -> "ebiten-game-jam"

**ClientID** is the ClientID of the OAuth Application created under the developer's/host's itch.io Settings.

**OAuthRedirect** is the OAuth redirect URL. This should be the location where the jam voting is accessible from, but with "auth" appended. e.g, "https://kettek.net/jams/ebiten-jam" -> "https://kettek.net/jams/ebiten-jam/auth"

**VotingEnabled** allows voting to be done by users that have logged in.

**VotingFinished** prevents any further voting and shows the averaged results on the entries index.

## Running
You can either build and run the go application or just issue `go run .`. This will read the config.json file, set up needed databases, collect information of the jam, then start the server.