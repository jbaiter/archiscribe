export function preloadLineImages (lines) {
  let preloadImage = (url, retryCount = 0) => {
    let img = new Image()
    img.src = url
    img.onerror = () => {
      if (retryCount < 5) {
        retryCount += 1
        setTimeout(preloadImage, retryCount * 1000, url, retryCount)
      }
    }
  }

  lines.forEach((url) => {
    preloadImage(url)
  })
}

export function getRandomInt (min, max) {
  min = Math.ceil(min)
  max = Math.floor(max)
  return Math.floor(Math.random() * (max - min)) + min
}
