export function renderReaderPage(mangaId: string, chapterNumber: number, pageNumber: number): void {
    const app = document.getElementById('app');
    if (!app) return;
    
    app.innerHTML = `
      <div class="reader-container">
        <h2>Reading Chapter ${chapterNumber}, Page ${pageNumber}</h2>
        <p>Loading page...</p>
        <div class="reader-image-container"></div>
      </div>
    `;
    
    // Load page will be implemented later
  }