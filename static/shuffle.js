window.addEventListener('DOMContentLoaded', () => {
  try {
    let order = localStorage.getItem("order")
    if (!order) {
      let entries = [...document.getElementsByClassName('entry')]
      for (let i = entries.length -1; i > 0; i--) {
        let j = Math.floor(Math.random() * (i+1))
        let t = entries[i]
        entries[i] = entries[j]
        entries[j] = t
      }
      order = entries.map(v=>v.dataset.id)
      localStorage.setItem("order", JSON.stringify(order))
    } else {
      order = JSON.parse(order)
      // TODO: We should probably check entries for any IDs not in the order array, and if so, to randomly insert them into the order array.
    }

    for (let entry of [...document.getElementsByClassName('entry')].sort((a,b) => order.indexOf(a.dataset.id) - order.indexOf(b.dataset.id))) {
      document.getElementsByClassName('entries')[0].appendChild(entry)
    }
  } catch (err) {
    // Eh... just silently ignore. This is just fluff and if localStorage doesn't exist, then it doesn't really matter.
  }
})