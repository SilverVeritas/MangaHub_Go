import { getMangaById, getChapters } from '../services/api';
import { navigateTo } from '../routes';
import { Chapter } from '../types/manga';

export async function renderMangaPage(mangaId: string): Promise<void> {
  const app = document.getElementById('app');
  if (!app) return;
  
  // Show loading state
  app.innerHTML = `
    <div class="manga-page">
      <div class="back-link">
        <a href="/" id="back-to-home">← Back to Home</a>
      </div>
      <p id="loading-message">Loading manga details...</p>
      <div class="manga-details"></div>
    </div>
  `;
  
  document.getElementById('back-to-home')?.addEventListener('click', (e) => {
    e.preventDefault();
    navigateTo('/');
  });
  
  try {
    // Fetch manga details and chapters in parallel
    const [manga, chapters] = await Promise.all([
      getMangaById(mangaId),
      getChapters(mangaId)
    ]);
    
    const detailsContainer = document.querySelector('.manga-details');
    const loadingMessage = document.getElementById('loading-message');
    
    if (loadingMessage) {
      loadingMessage.style.display = 'none';
    }
    
    if (detailsContainer) {
      // Sort chapters in descending order (newest first)
      const sortedChapters = [...chapters].sort((a, b) => b.number - a.number);
      
      detailsContainer.innerHTML = `
        <div class="manga-header">
          <div class="manga-cover-large">
            <img src="${manga.coverImage}" alt="${manga.title}">
          </div>
          <div class="manga-info-detailed">
            <h1>${manga.title}</h1>
            ${manga.altTitles?.length ? `<div class="alt-titles">Also known as: ${manga.altTitles.join(', ')}</div>` : ''}
            <div class="manga-metadata">
              <span class="author">Author: ${manga.author}</span>
              ${manga.artist ? `<span class="artist">Artist: ${manga.artist}</span>` : ''}
              <span class="status">Status: ${manga.status}</span>
              ${manga.publishedYear ? `<span class="year">Year: ${manga.publishedYear}</span>` : ''}
            </div>
            <div class="genres">
              ${manga.genres.map(genre => `<span class="genre-tag">${genre}</span>`).join('')}
            </div>
            <div class="description">
              <h3>Description</h3>
              <p>${manga.description || 'No description available.'}</p>
            </div>
          </div>
        </div>
        
        <div class="chapters-section">
          <h2>Chapters</h2>
          <div class="chapters-list">
            ${renderChaptersList(sortedChapters, mangaId)}
          </div>
        </div>
      `;
      
      // Add event listeners to chapter items
      document.querySelectorAll('.chapter-item').forEach(item => {
        item.addEventListener('click', (e) => {
          e.preventDefault();
          const chapterNumber = (item as HTMLElement).dataset.number;
          if (chapterNumber) {
            navigateTo(`/reader/${mangaId}/${chapterNumber}/1`);
          }
        });
      });
    }
  } catch (error) {
    console.error('Failed to load manga details:', error);
    const detailsContainer = document.querySelector('.manga-details');
    if (detailsContainer) {
      detailsContainer.innerHTML = `
        <div class="error-state">
          <p>Failed to load manga details. Please try again later.</p>
          <button id="retry-button">Retry</button>
        </div>
      `;
      
      document.getElementById('retry-button')?.addEventListener('click', () => {
        renderMangaPage(mangaId);
      });
    }
  }
}

// Helper function to render the chapters list
function renderChaptersList(chapters: Chapter[], mangaId: string): string {
  if (chapters.length === 0) {
    return '<p class="no-chapters">No chapters available.</p>';
  }
  
  return `
    <ul class="chapters-list-items">
      ${chapters.map(chapter => {
        const releaseDate = new Date(chapter.releaseDate);
        const formattedDate = releaseDate.toLocaleDateString();
        
        return `
          <li class="chapter-item" data-number="${chapter.number}">
            <div class="chapter-info">
              <span class="chapter-number">Chapter ${chapter.number}</span>
              <span class="chapter-title">${chapter.title || ''}</span>
            </div>
            <div class="chapter-meta">
              <span class="chapter-date">${formattedDate}</span>
              <span class="chapter-pages">${chapter.pageCount} pages</span>
            </div>
          </li>
        `;
      }).join('')}
    </ul>
  `;
}