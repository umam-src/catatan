import adapter from '@sveltejs/adapter-static';

export default {
  kit: {
    adapter: adapter({ pages: '../internal/ui/dist', assets: '../internal/ui/dist', fallback: 'index.html' })
  }
};
