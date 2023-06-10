package main

import (
	"html/template"
)

func loadTemplates() *template.Template {
	template, err := template.New("").Funcs(template.FuncMap{
		"getFinalVotes": func(id int) Votes {
			votes, _ := getFinalVotes(id)
			return votes
		},
		"getVotes": func(user UserDetails, id int) Votes {
			votes, _ := getVotes(user, id)
			return votes
		},
		"iterate": func(count int) []int {
			var stars []int
			for i := 0; i < count; i++ {
				stars = append(stars, i+1)
			}
			return stars
		},
		"starSet": func(cat float64, v int) bool {
			return cat >= float64(v)
		},
		"isOwnGame": func(user UserDetails, id int) bool {
			return isOwnGame(user, id)
		},
	}).ParseFiles("templates/index.gohtml", "templates/entry.gohtml", "templates/auth.gohtml", "templates/header.gohtml")
	if err != nil {
		panic(err)
	}
	return template
}
