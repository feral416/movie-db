package movie

type Context struct {
	Movie  *Movie
	Genres []string
}

type MovieActionsContext struct {
	*Movie
	Error string
	Msg   string
}

func NewMovieActionsContext() *MovieActionsContext {
	return &MovieActionsContext{Movie: &Movie{}}
}

type ContextAddMovie struct {
	Movie  *Movie
	Errors []string
	ID     int64
}

func newEmptyContextAddMovie() (c *ContextAddMovie) {
	return &ContextAddMovie{
		&Movie{},
		[]string{},
		0,
	}
}

type MovContext struct {
	Movie
	Last bool
}
