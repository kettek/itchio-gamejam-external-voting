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
		"getBadges": func(user UserDetails, id int) Badges {
			badges, _ := getBadges(user, id)
			return badges
		},
		"getFinalBadges": func() BadgeResults {
			return getFinalBadges()
		},
		"getActualFinalBadges": func(results BadgeResults, id int) map[string]bool {
			badges := make(map[string]bool)
			for k, v := range results {
				for _, v2 := range v {
					if v2 == id {
						badges[k] = true
					}
				}
			}
			return badges
		},
		"getTags": func(user UserDetails, id int) Tags {
			tags, _ := getTags(user, id)
			return tags
		},
		"getFinalTags": func() TagResults {
			return getFinalTags()
		},
		"getActualFinalTags": func(results TagResults, id int) Tags {
			tags := make(Tags)
			for k, v := range results {
				for _, v2 := range v {
					if v2 == id {
						tags[k] = true
					}
				}
			}
			return tags
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
