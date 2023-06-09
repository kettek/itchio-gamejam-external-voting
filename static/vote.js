window.addEventListener('DOMContentLoaded', () => {
  const setGameVotes = (o) => {
    let entries = document.getElementsByClassName('entry')
    for (let entry of entries) {
      if (entry.dataset['id'] == o.id) {
        for (let category of entry.getElementsByClassName('ratings__entry__stars')) {
          if (!o.votes[category.dataset['category']]) continue
          let anchors = category.getElementsByTagName('a')
          for (let i = 0; i < anchors.length; i++) {
            if (i < o.votes[category.dataset['category']]) {
              anchors[i].innerHTML = '★'
            } else {
              anchors[i].innerHTML = '☆'
            }
          }
        }
        for (let number of entry.getElementsByClassName('ratings__entry__number')) {
          if (!o.votes[number.dataset['category']]) continue
          number.innerHTML = o.votes[number.dataset['category']]
        }
        return
      }
    }
  }

  let ratingsEntries = document.body.getElementsByClassName('ratings__entry__stars')
  for (let entry of ratingsEntries) {
    let links = entry.getElementsByTagName('a')
    for (let l of links) {
      l.addEventListener('click', async e => {
        e.preventDefault()
        const res = await fetch(l.href)
        if (res.status == 200) {
          const votes = await res.json()
          setGameVotes(votes)
        } else {
          console.log(res.status, await res.text())
        }
      })
    }
  }
})