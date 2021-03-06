package main

import (
	"html/template"
)

var (
	indexTemplate = template.Must(template.New("htmlTemplate").Parse(`
{{ $SelectedGuildID := .SelectedGuildID }}
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
    <link href="dashboard.css" rel="stylesheet">

    <title>Who's streaming right now?</title>
  </head>
  <body>
    <nav class="navbar navbar-dark fixed-top bg-dark flex-md-nowrap p-0 shadow">
      <a class="navbar-brand col-sm-3 col-md-2 mr-0" href="#">Who's streaming right now?</a>
      <ul class="navbar-nav px-3">
        <li class="nav-item text-nowrap">
          <a class="nav-link" href="#">Sign out</a>
        </li>
      </ul>
    </nav>

    <div class="container-fluid">
      <div class="row">
        <nav class="col-md-2 d-none d-md-block bg-light sidebar">
          <div class="sidebar-sticky">
            <ul class="nav flex-column">
              {{range $idx, $guild := .Guilds}}
              <li class="nav-item">
                <a class="nav-link {{ if eq $guild.ID $SelectedGuildID }}active{{end}}" href="?guild={{ $guild.ID }}">
                  {{ $guild.Name }}
                  {{ if eq $guild.ID $SelectedGuildID }}
                  <span class="sr-only">(current)</span>
                  {{ end }}
                </a>
              </li>
              {{ end }}
            </ul>
            <div class="dropdown-divider"></div>
            <p>
              Don't see your server?
              <br />
              <a href="{{ .BotAddURL }}">Add the bot</a>
            </p>
          </div>
        </nav>

        <main role="main" class="col-md-9 ml-sm-auto col-lg-10 px-4">
          <h1>Streamers</h1>
          <div class="container">
            <div class="row">
              {{range $idx, $stream := .TwitchStreams}}
              <div class="col">
                <div id="stream{{ $idx }}"></div>
                <div><a href="{{ $stream.URL }}">{{ $stream.Channel }}</a></div>
              </div>
              {{ else }}
              <p>
                None of the streamers for this server are active.
                <br />
                Couldn't find your stream?
                <br />
                Make sure you've added it by running the following command in discord
                <br />
                <kbd>!addTwitch https://www.twitch.tv/yourusername</kbd>
              </p>
              {{end}}
            </div>
          </div>

        </main>
      </div>
    </div>

    <!-- Optional JavaScript -->
    <!-- jQuery first, then Popper.js, then Bootstrap JS -->
    <script src="https://code.jquery.com/jquery-3.2.1.slim.min.js" integrity="sha384-KJ3o2DKtIkvYIK3UENzmM7KCkRr/rE9/Qpg6aAZGJwFDMVNA/GpGFF93hXpG5KkN" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.9/umd/popper.min.js" integrity="sha384-ApNbgh9B+Y1QKtv3Rn7W3mgPxhU9K/ScQsAP7hUibX39j7fakFPskvXusvfa0b4Q" crossorigin="anonymous"></script>
    <script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js" integrity="sha384-JZR6Spejh4U02d8jOt6vLEHfe/JQGiRRSQQxSfFWpi1MquVdAyjUar5+76PVCmYl" crossorigin="anonymous"></script>

    <!-- Load the Twitch embed script -->
    <script src="https://embed.twitch.tv/embed/v1.js"></script>
    <!-- Create a Twitch.Embed object that will render within the "twitch-embed" root element. -->
    <script type="text/javascript">
      {{range $idx, $stream := .TwitchStreams}}
      new Twitch.Embed("stream{{ $idx}}", {
        width: 427,
        height: 240,
        channel: "{{ $stream.Channel }}",
        layout: "video"
      });
      {{end}}
    </script>
  </body>
</html>`))
)
