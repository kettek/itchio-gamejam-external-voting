window.addEventListener('DOMContentLoaded', () => {
  let admin = document.getElementsByClassName('admin')[0]

  const setAdminConfig = (o) => {
    for (let input of admin.getElementsByTagName('input')) {
      if (o[input.name] !== undefined) {
        if (input.type === 'checkbox') {
          input.checked = o[input.name] ? true : false
        } else {
          input.value = o[input.name]
        }
      }
    }
    // It's easier just to recreate the vote category elements than update them.
    for (let el of [...document.getElementsByClassName('voteCategory')]) {
      el.parentElement.removeChild(el)
    }
    for (let el of [...document.getElementsByClassName('badge')]) {
      el.parentElement.removeChild(el)
    }
    for (let el of [...document.getElementsByClassName('tag')]) {
      el.parentElement.removeChild(el)
    }
    for (let i = 0; i < o.VoteCategories.length; i++) {
      createVoteCategory('VoteCategories-'+i, o.VoteCategories[i])
    }
    for (let i = 0; i < o.Badges.length; i++) {
      createBadge('Badges-'+i, o.Badges[i])
    }
    for (let i = 0; i < o.Tags.length; i++) {
      createTag('Tags-'+i, o.Tags[i])
    }
  }
  
  const createVoteCategory = (name, value) => {
    let span = document.createElement('span')
    span.className = 'voteCategory'
    let input = document.createElement('input')
    input.className = 'VoteCategories'
    input.name = name
    input.value = value
    setupElement(input)
    let button = document.createElement('button')
    button.className = 'RemoveVoteCategory'
    button.innerHTML = 'Remove'
    setupRemoveCategory(button)
    span.appendChild(input)
    span.appendChild(button)
    
    document.getElementById('NewVoteCategory').parentElement.parentElement.insertBefore(
      span,
      document.getElementById('NewVoteCategory').parentElement
    )
  }
  
  const createBadge = (name, value) => {
    let span = document.createElement('span')
    span.className = 'badge'
    let input = document.createElement('input')
    input.className = 'Badges'
    input.name = name
    input.value = value
    setupElement(input)
    let button = document.createElement('button')
    button.className = 'RemoveBadge'
    button.innerHTML = 'Remove Badge'
    setupRemoveBadge(button)
    span.appendChild(input)
    span.appendChild(button)
    
    document.getElementById('NewBadge').parentElement.parentElement.insertBefore(
      span,
      document.getElementById('NewBadge').parentElement
    )
  }

  const createTag = (name, value) => {
    let span = document.createElement('span')
    span.className = 'tag'
    let input = document.createElement('input')
    input.className = 'Tag'
    input.name = name
    input.value = value
    setupElement(input)
    let button = document.createElement('button')
    button.className = 'RemoveTag'
    button.innerHTML = 'Remove Tag'
    setupRemoveTag(button)
    span.appendChild(input)
    span.appendChild(button)
    
    document.getElementById('NewTag').parentElement.parentElement.insertBefore(
      span,
      document.getElementById('NewTag').parentElement
    )
  }
  
  const setupElement = (el) => {
    el.addEventListener('change', async e => {
      e.stopPropagation()
      e.preventDefault()
      let res
      if (e.currentTarget.getAttribute('type') === 'checkbox') {
        res = await fetch(`admin?${e.currentTarget.getAttribute('name')}=${e.currentTarget.checked}`)
      } else {
        res = await fetch(`admin?${e.currentTarget.getAttribute('name')}=${e.currentTarget.value}`)
      }
      if (res.status === 200) {
        const conf = await res.json()
        setAdminConfig(conf)
      } else {
        console.log(res.status, await res.text())
      }
    })
  }
  
  const setupRemoveCategory = (el) => {
    el.addEventListener('click', async e => {
      let res = await fetch(`admin?RemoveVoteCategory=${e.currentTarget.previousElementSibling.name}`)
      if (res.status === 200) {
        el.parentElement.parentElement.removeChild(el.parentElement)
        const conf = await res.json()
        setAdminConfig(conf)
      } else {
        console.log(res.status, await res.text())
      }
    })
  }
  
  document.getElementById('AddVoteCategory')?.addEventListener('click', async e => {
    let res = await fetch(`admin?AddVoteCategory=${document.getElementById('NewVoteCategory').value}`)
    if (res.status === 200) {
      document.getElementById('NewVoteCategory').value = ''
      
      const conf = await res.json()
      setAdminConfig(conf)
    } else {
      console.log(res.status, await res.text())
    }
  })
  
  for (let button of admin.getElementsByClassName('RemoveVoteCategory')) {
    setupRemoveCategory(button)
  }
  
  for (let input of admin.getElementsByTagName('input')) {
    if (input.name === 'NewVoteCategory' || input.name === 'NewBadge' || input.name === 'NewTag') {
      continue
    }
    setupElement(input)
  }
  
  const setupRemoveBadge = (el) => {
    el.addEventListener('click', async e => {
      let res = await fetch(`admin?RemoveBadge=${e.currentTarget.previousElementSibling.name}`)
      if (res.status === 200) {
        el.parentElement.parentElement.removeChild(el.parentElement)
        const conf = await res.json()
        setAdminConfig(conf)
      } else {
        console.log(res.status, await res.text())
      }
    })
  }

  document.getElementById('AddBadge')?.addEventListener('click', async e => {
    let res = await fetch(`admin?AddBadge=${document.getElementById('NewBadge').value}`)
    if (res.status === 200) {
      document.getElementById('NewBadge').value = ''
      
      const conf = await res.json()
      setAdminConfig(conf)
    } else {
      console.log(res.status, await res.text())
    }
  })

  for (let button of admin.getElementsByClassName('RemoveBadge')) {
    setupRemoveBadge(button)
  }

  const setupRemoveTag = (el) => {
    el.addEventListener('click', async e => {
      let res = await fetch(`admin?RemoveTag=${e.currentTarget.previousElementSibling.name}`)
      if (res.status === 200) {
        el.parentElement.parentElement.removeChild(el.parentElement)
        const conf = await res.json()
        setAdminConfig(conf)
      } else {
        console.log(res.status, await res.text())
      }
    })
  }

  document.getElementById('AddTag')?.addEventListener('click', async e => {
    let res = await fetch(`admin?AddTag=${document.getElementById('NewTag').value}`)
    if (res.status === 200) {
      document.getElementById('NewTag').value = ''
      
      const conf = await res.json()
      setAdminConfig(conf)
    } else {
      console.log(res.status, await res.text())
    }
  })

  for (let button of admin.getElementsByClassName('RemoveTag')) {
    setupRemoveTag(button)
  }
})