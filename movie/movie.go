package movie

import "strings"

type Movie struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Genres string //`json:"genres"`
}

func (m *Movie) SplitGenresString() []string {
	return strings.Split(m.Genres, "|")
}

/*func (m *Movie) PackGenresToString(elems []string) string {
	return strings.Join(elems, "|")
}*/
