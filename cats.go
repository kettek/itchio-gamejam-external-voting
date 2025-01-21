package main

type CatResult struct {
	id    int
	value float64
	cat   string
}

type CatResults []CatResult

func (cr CatResults) Len() int           { return len(cr) }
func (cr CatResults) Swap(i, j int)      { cr[i], cr[j] = cr[j], cr[i] }
func (cr CatResults) Less(i, j int) bool { return cr[i].value > cr[j].value }

func getTopCats() []CatResult {
	top := make([]CatResult, 0)
	for _, e := range entries.Games {
		votes, err := getFinalVotes(e.ID)
		if err != nil {
			continue
		}
		for k, v := range votes {
			found := false
			for ci, ce := range top {
				if ce.cat == k {
					if ce.value < v {
						top[ci] = CatResult{
							id:    e.ID,
							value: v,
							cat:   k,
						}
					}
					found = true
					break
				}
			}
			if !found {
				top = append(top, CatResult{
					id:    e.ID,
					value: v,
					cat:   k,
				})
			}
		}
	}

	return top
}
