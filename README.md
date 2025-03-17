# Movie database

A website to manage database of movies written with HTMX and Go(standard library). Has basic functionality to manage movies, users, comments, basic admin functions. For data storage MySQL db is used with simple caching using concurent-safe map.

## Installation

1. `cd` to project folder
1. Execute `go build .`
1. Execute `go run .`
1. Install MySQL server
1. Execute script `/mysql/db.sql`
1. Create MySQL user for a site
1. Set environmental variables: `"MOVIE_DB_USER", "MOVIE_DB_PWD", "MYSQL_DB_ADDR"`
    corresponding to your database settings
1. Register user on the site and set `admin` flag for the your user in `users`table,
    refresh the page and you are ready to go!
