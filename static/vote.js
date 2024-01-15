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
  
  const setGameTags = (o) => {
    let entries = document.getElementsByClassName('entry')
    for (let entry of entries) {
      if (entry.dataset['id'] == o.id) {
        let tags = entry.getElementsByClassName('tags')[0]
        for (let tag of tags.getElementsByClassName('tags__entry')) {
          if (o.tags[tag.dataset['tag']]) {
            tag.classList.add('-selected')
          } else {
            tag.classList.remove('-selected')
          }
        }
        return
      }
    }
  }
  
  let tagEntries = document.body.getElementsByClassName('tags__entry')
  for (let entry of tagEntries) {
    let links = entry.getElementsByTagName('a')
    for (let l of links) {
      l.addEventListener('click', async e => {
        e.preventDefault()
        const res = await fetch(l.href)
        if (res.status == 200) {
          const tags = await res.json()
          setGameTags(tags)
        } else {
          console.log(res.status, await res.text())
        }
      })
    }
  }
  
  const setBadges = (o) => {
    let entries = document.getElementsByClassName('entry')
    for (let entry of entries) {
      let processed = Object.entries(o).length
      for (let [id, badges] of Object.entries(o)) {
        if (id !== entry.dataset['id']) continue
        processed--
        let badgesContainer = entry.getElementsByClassName('badges')[0]
        for (let badge of badgesContainer.getElementsByClassName('badges__entry')) {
          if (badges[badge.dataset['badge']]) {
            badge.classList.add('-selected')
          } else {
            badge.classList.remove('-selected')
          }
        }
      }
      if (processed === 0) return
    }
  }
  
  let badgeEntries = document.body.getElementsByClassName('badges__entry')
  for (let entry of badgeEntries) {
    let links = entry.getElementsByTagName('a')
    for (let l of links) {
      l.addEventListener('click', async e => {
        e.preventDefault()
        const res = await fetch(l.href)
        if (res.status == 200) {
          const badges = await res.json()
          setBadges(badges)
        } else {
          console.log(res.status, await res.text())
        }
      })
    }
  }

})