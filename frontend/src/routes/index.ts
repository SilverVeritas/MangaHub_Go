import { renderHomePage } from '../components/home-page';
import { renderMangaPage } from '../components/manga-page';
import { renderReaderPage } from '../components/reader-page';

export function initRouter() {
  // Handle navigation events
  window.addEventListener('popstate', handleRouteChange);
  
  // Handle clicks on links
  document.addEventListener('click', (e) => {
    const target = e.target as HTMLElement;
    const link = target.closest('a');
    
    if (link && link.href.startsWith(window.location.origin) && !link.dataset.external) {
      e.preventDefault();
      navigateTo(link.href);
    }
  });
  
  // Initial route
  handleRouteChange();
}

export function navigateTo(url: string) {
  history.pushState(null, '', url);
  handleRouteChange();
}

function handleRouteChange() {
  const path = window.location.pathname;
  
  // Match routes
  if (path === '/' || path === '/index.html') {
    renderHomePage();
  } else if (path.match(/^\/manga\/[a-zA-Z0-9-]+\/?$/)) {
    const mangaId = path.split('/')[2];
    renderMangaPage(mangaId);
  } else if (path.match(/^\/reader\/[a-zA-Z0-9-]+\/[\d.]+\/\d+\/?$/)) {
    const [, , mangaId, chapterNumber, pageNumber] = path.split('/');
    renderReaderPage(mangaId, parseFloat(chapterNumber), parseInt(pageNumber, 10));
  } else {
    // 404 page
    renderHomePage(); // Or a dedicated 404 page
  }
}