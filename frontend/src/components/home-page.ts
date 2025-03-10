import { getAllManga } from '../services/api';
import { Manga } from '../types/manga';
import { navigateTo } from '../routes';

export async function renderHomePage(): Promise<void> {
  const app = document.getElementById('app');
  if (!app) return;
  
  // Show loading state
  app.innerHTML = `
    <header>
      <h1>MangaHub</h1>
    </header>
    <main>
      <p id="loading-message">Loading manga collection...</p>
      <div class="manga-grid"></div>
    </main>
  `;
  
  try {
    // Fetch manga collection
    const mangaCollection = await getAllManga();
    const mangaGrid = document.querySelector('.manga-grid');
    const loadingMessage = document.getElementById('loading-message');
    
    if (loadingMessage) {
      loadingMessage.style.display = 'none';
    }
    
    if (mangaGrid && mangaCollection.length > 0) {
      // Render each manga card
      mangaGrid.innerHTML = mangaCollection.map(manga => createMangaCard(manga)).join('');
      
      // Add event listeners to manga cards
      document.querySelectorAll('.manga-card').forEach(card => {
        card.addEventListener('click', (e) => {
          const mangaId = (card as HTMLElement).dataset.id;
          if (mangaId) {
            e.preventDefault();
            navigateTo(`/manga/${mangaId}`);
          }
        });
      });
    } else if (mangaGrid) {
      // No manga found
      mangaGrid.innerHTML = `
        <div class="empty-state">
          <p>No manga found. Add some manga to your collection!</p>
        </div>
      `;
    }
  } catch (error) {
    console.error('Failed to load manga collection:', error);
    const mangaGrid = document.querySelector('.manga-grid');
    if (mangaGrid) {
      mangaGrid.innerHTML = `
        <div class="error-state">
          <p>Failed to load manga collection. Please try again later.</p>
          <button id="retry-button">Retry</button>
        </div>
      `;
      
      document.getElementById('retry-button')?.addEventListener('click', () => {
        renderHomePage();
      });
    }
  }
}

// Helper function to create a manga card
function createMangaCard(manga: Manga): string {
  return `
    <div class="manga-card" data-id="${manga.id}">
      <div class="manga-cover-container">
        <img class="manga-cover" src="${manga.coverImage}" alt="${manga.title}" loading="lazy">
      </div>
      <div class="manga-info">
        <h3 class="manga-title">${manga.title}</h3>
        <div class="manga-meta">
          <span class="manga-chapters">${manga.chapterCount} chapter${manga.chapterCount !== 1 ? 's' : ''}</span>
          <span class="manga-status">${manga.status}</span>
        </div>
      </div>
    </div>
  `;
}