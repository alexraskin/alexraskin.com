document.addEventListener('DOMContentLoaded', function() {
    // Load LastFM data on page load
    loadLastFM();
    
    // Refresh LastFM data every minute
    setInterval(loadLastFM, 60000);
});

async function loadLastFM() {
    try {
        const response = await fetch(`/api/lastfm`, {
            method: "GET"
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        const html = await response.text();
        
        document.querySelector("#lastfm").innerHTML = html;
    } catch (e) {
        console.error("Error fetching LastFM data:", e);
        lastFMError();
    }
}

function lastFMError() {
    document.querySelector("#lastfm").innerHTML = `<span class="error">Error fetching last.fm data</span>`;
} 