window.addEventListener("load", () => {
  const queryString = window.location.hash.slice(1)
  const params = new URLSearchParams(queryString)
  const accessToken = params.get("access_token")
  if (accessToken) {
    window.location.replace(window.location.href.substring(0, window.location.href.lastIndexOf('/'))+'/'+'auth'+'?key='+accessToken)
  }
})