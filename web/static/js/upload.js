document.addEventListener('DOMContentLoaded', () => {
  const input = document.getElementById('upload-input');
  if (!input) return;

  input.addEventListener('change', async () => {
    const file = input.files && input.files[0];
    if (!file) return;

    const base = window.currentListPath === '/' ? '' : window.currentListPath;
    const target = base ? `${base}/${file.name}` : file.name;
    const form = new FormData();
    form.append('path', target);
    form.append('file', file);

    const response = await fetch('/upload', {
      method: 'POST',
      body: form,
    });

    if (!response.ok) {
      console.error('upload failed');
      return;
    }

    input.value = '';
    if (typeof loadFiles === 'function') {
      await loadFiles(window.currentListPath || '/');
    }
  });
});
