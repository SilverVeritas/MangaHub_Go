export function renderHomePage(): void {
    const app = document.getElementById('app');
    if (!app) return;
    
    app.innerHTML = `
      <h1>MangaHub</h1>
      <p>Loading manga collection...</p>
      <div class="manga-grid"></div>
    `;
    
    // Load manga will be implemented later
  }