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
  }

  for (let input of admin.getElementsByTagName('input')) {
    input.addEventListener('change', async e => {
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
})