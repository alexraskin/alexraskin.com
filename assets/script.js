document.addEventListener('DOMContentLoaded', () => {
  updateNowPlaying();
});

function updateNowPlaying() {
  const nowPlayingElement = document.getElementById('now-playing');
  const refreshIcon = document.querySelector('.refresh-icon');
  
  if (refreshIcon) {
    refreshIcon.classList.add('rotating');
  }
  nowPlayingElement.textContent = "Loading...";
  
  fetch('https://lastfm.alexraskin.com/twizycat', {
    method: 'GET',
    headers: {
      'Accept': 'application/json'
    },
    timeout: 5000
  })
  .then(response => {
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    return response.json();
  })
  .then(data => {
    if (data && (data.track && data.artist)) {
      nowPlayingElement.textContent = `"${data.track} by ${data.artist}"`;
    } else {
      nowPlayingElement.textContent = `"Not listening to anything"`;
    }
  })
  .catch(error => {
    console.error('Error fetching now playing data:', error);
    nowPlayingElement.textContent = `"Unable to fetch music data"`;
  })
  .finally(() => {
    if (refreshIcon) {
      refreshIcon.classList.remove('rotating');
    }
  });
}
