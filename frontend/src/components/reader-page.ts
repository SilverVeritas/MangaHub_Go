import { getPage, getChapters, getMangaById } from '../services/api';
import { navigateTo } from '../routes';

export async function renderReaderPage(mangaId: string, chapterNumber: number, pageNumber: number): Promise<void> {
  const app = document.getElementById('app');
  if (!app) return;
  
  // Show loading state
  app.innerHTML = `
    <div class="reader-container">
      <div class="reader-header">
        <a href="/manga/${mangaId}" class="back-to-manga">← Back to Details</a>
        <h2>Chapter ${chapterNumber}</h2>
        <div class="reader-page-info">Loading...</div>
      </div>
      <div class="reader-loading">Loading page...</div>
      <div class="reader-image-container">
        <!-- Page will be displayed here -->
      </div>
      <div class="reader-controls">
        <button id="prev-page" class="reader-nav-button">&lt; Previous</button>
        <div class="page-selector">
          <select id="page-select" disabled>
            <option>Loading...</option>
          </select>
        </div>
        <button id="next-page" class="reader-nav-button">Next &gt;</button>
      </div>
    </div>
  `;
  
  try {
    // Fetch page data and manga info
    const pageData = await getPage(mangaId, chapterNumber, pageNumber);
    const manga = await getMangaById(mangaId);
    const chapters = await getChapters(mangaId);
    
    // Update the page info
    const pageInfoElement = document.querySelector('.reader-page-info');
    if (pageInfoElement) {
      pageInfoElement.textContent = `Page ${pageNumber} of ${pageData.totalPages || '?'}`;
    }
    
    // Find the current chapter for title
    const currentChapter = chapters.find(ch => ch.number === chapterNumber);
    const chapterTitle = document.querySelector('.reader-header h2');
    if (chapterTitle && currentChapter) {
      chapterTitle.textContent = `Chapter ${chapterNumber}${currentChapter.title ? ': ' + currentChapter.title : ''}`;
    }
    
    // Hide loading message
    const loadingElement = document.querySelector('.reader-loading');
    if (loadingElement) {
      (loadingElement as HTMLElement).style.display = 'none';
    }
    
    // Display the page image
    const imageContainer = document.querySelector('.reader-image-container');
    if (imageContainer) {
      imageContainer.innerHTML = `
        <img src="${pageData.imageUrl}" alt="Page ${pageNumber}" class="reader-image">
      `;
      
      // Add click navigation
      const imageElement = imageContainer.querySelector('img');
      if (imageElement) {
        imageElement.addEventListener('click', (e) => {
          const rect = imageElement.getBoundingClientRect();
          const clickX = e.clientX - rect.left;
          
          // Click on left third of image goes to previous page, right third goes to next page
          if (clickX < rect.width / 3) {
            navigateToPrevPage(mangaId, chapterNumber, pageNumber, pageData);
          } else if (clickX > (rect.width * 2 / 3)) {
            navigateToNextPage(mangaId, chapterNumber, pageNumber, pageData);
          }
          // Middle third does nothing, letting users see the menu if they click there
        });
      }
    }
    
    // Update page selector
    const pageSelector = document.getElementById('page-select') as HTMLSelectElement;
    if (pageSelector && pageData.totalPages) {
      pageSelector.innerHTML = Array.from({ length: pageData.totalPages }, (_, i) => {
        const pageNum = i + 1;
        return `<option value="${pageNum}" ${pageNum === pageNumber ? 'selected' : ''}>Page ${pageNum}</option>`;
      }).join('');
      pageSelector.disabled = false;
      
      pageSelector.addEventListener('change', () => {
        const newPage = parseInt(pageSelector.value, 10);
        navigateTo(`/reader/${mangaId}/${chapterNumber}/${newPage}`);
      });
    }
    
    // Setup navigation buttons
    const prevButton = document.getElementById('prev-page') as HTMLButtonElement;
    const nextButton = document.getElementById('next-page') as HTMLButtonElement;
    
    if (prevButton) {
      if (pageNumber <= 1 && !pageData.prevChapter) {
        prevButton.disabled = true;
      } else {
        prevButton.addEventListener('click', () => {
          navigateToPrevPage(mangaId, chapterNumber, pageNumber, pageData);
        });
      }
    }
    
    if (nextButton) {
      if (pageNumber >= (pageData.totalPages || 0) && !pageData.nextChapter) {
        nextButton.disabled = true;
      } else {
        nextButton.addEventListener('click', () => {
          navigateToNextPage(mangaId, chapterNumber, pageNumber, pageData);
        });
      }
    }
    
    // Set up keyboard navigation
    document.addEventListener('keydown', handleKeyNavigation);
    
    // Cleanup function to remove the event listener when leaving the page
    window.addEventListener('popstate', () => {
      document.removeEventListener('keydown', handleKeyNavigation);
    });
    
    function handleKeyNavigation(e: KeyboardEvent) {
      if (e.key === 'ArrowLeft' || e.key === 'a') {
        navigateToPrevPage(mangaId, chapterNumber, pageNumber, pageData);
      } else if (e.key === 'ArrowRight' || e.key === 'd') {
        navigateToNextPage(mangaId, chapterNumber, pageNumber, pageData);
      }
    }
    
  } catch (error) {
    console.error('Failed to load page:', error);
    const imageContainer = document.querySelector('.reader-image-container');
    if (imageContainer) {
      imageContainer.innerHTML = `
        <div class="error-state">
          <p>Failed to load page. Please try again later.</p>
          <button id="retry-button">Retry</button>
        </div>
      `;
      
      document.getElementById('retry-button')?.addEventListener('click', () => {
        renderReaderPage(mangaId, chapterNumber, pageNumber);
      });
    }
  }
}

// Helper function to navigate to the previous page
function navigateToPrevPage(
  mangaId: string, 
  chapterNumber: number, 
  pageNumber: number, 
  pageData: any
) {
  if (pageNumber > 1) {
    // Go to previous page in same chapter
    navigateTo(`/reader/${mangaId}/${chapterNumber}/${pageNumber - 1}`);
  } else if (pageData.prevChapter) {
    // Go to last page of previous chapter
    getChapters(mangaId)
      .then(chapters => {
        const prevChapter = chapters.find(ch => ch.number === parseFloat(pageData.prevChapter));
        if (prevChapter) {
          navigateTo(`/reader/${mangaId}/${prevChapter.number}/${prevChapter.pageCount}`);
        }
      })
      .catch(error => {
        console.error('Failed to navigate to previous chapter:', error);
      });
  }
}

// Helper function to navigate to the next page
function navigateToNextPage(
  mangaId: string, 
  chapterNumber: number, 
  pageNumber: number, 
  pageData: any
) {
  if (pageNumber < (pageData.totalPages || 0)) {
    // Go to next page in same chapter
    navigateTo(`/reader/${mangaId}/${chapterNumber}/${pageNumber + 1}`);
  } else if (pageData.nextChapter) {
    // Go to first page of next chapter
    navigateTo(`/reader/${mangaId}/${pageData.nextChapter}/1`);
  }
}