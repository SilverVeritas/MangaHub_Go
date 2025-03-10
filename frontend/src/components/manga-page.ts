export function renderMangaPage(mangaId: string): void {
    const app = document.getElementById('app');
    if (!app) return;
    
    app.innerHTML = `
      <h1>Manga Details</h1>
      <p>Loading manga ${mangaId}...</p>
      <div class="manga-details"></div>
    `;
    
    // Load manga details will be implemented later
  }