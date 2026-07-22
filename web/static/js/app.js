window.currentListPath = '/';
window.workspaceView = 'files';

function setWorkspaceView(view) {
  const shell = document.querySelector('.product-shell');
  const editor = document.getElementById('code-window');
  if (!shell || !editor) return;

  const hasFile = Boolean(window.currentEditorPath);
  const resolvedView = !hasFile && view !== 'files' ? 'files' : view;
  window.workspaceView = resolvedView;
  shell.classList.remove('view-files', 'view-split', 'view-editor');
  shell.classList.add(`view-${resolvedView}`);
  editor.classList.toggle('is-hidden', resolvedView === 'files' || !hasFile);

  document.querySelectorAll('.view-button').forEach((button) => {
    button.classList.toggle('is-active', button.dataset.view === resolvedView);
    button.disabled = !hasFile && button.dataset.view !== 'files';
  });

  if (window.codeEditor && resolvedView !== 'files') {
    window.setTimeout(() => window.codeEditor.refresh(), 0);
  }
}

function closeEditor() {
  window.currentEditorPath = '';
  const currentPath = document.getElementById('current-path');
  const saveButton = document.getElementById('save-btn');
  const saveStatus = document.getElementById('save-status');
  if (currentPath) currentPath.textContent = 'Select a file to begin';
  if (saveButton) saveButton.disabled = true;
  if (saveStatus) saveStatus.textContent = '';
  setWorkspaceView('files');
}

async function apiJSON(url, options = {}) {
  const response = await fetch(url, options);
  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `request failed: ${response.status}`);
  }
  return response.json();
}

function renderBreadcrumbs(path) {
  const container = document.getElementById('browse-path');
  if (!container) return;
  container.innerHTML = '';

  const root = document.createElement('button');
  root.type = 'button';
  root.textContent = '/';
  root.addEventListener('click', () => loadFiles('/'));
  container.appendChild(root);

  const parts = path.split('/').filter(Boolean);
  parts.forEach((part, index) => {
    if (index > 0) {
      const separator = document.createElement('span');
      separator.textContent = '/';
      container.appendChild(separator);
    }

    const crumb = document.createElement('button');
    crumb.type = 'button';
    crumb.textContent = part;
    crumb.addEventListener('click', () => loadFiles(parts.slice(0, index + 1).join('/')));
    container.appendChild(crumb);
  });
}

async function loadFiles(path = '/') {
  window.currentListPath = path;
  const data = await apiJSON(`/files?path=${encodeURIComponent(path)}`);
  const list = document.getElementById('file-list');
  renderBreadcrumbs(path);
  if (!list) return;
  list.innerHTML = '';

  if (path !== '/') {
    const back = document.createElement('li');
    const button = document.createElement('button');
    button.type = 'button';
    button.textContent = '..';
    button.className = 'parent-row';
    button.addEventListener('click', () => {
      const parts = path.split('/').filter(Boolean);
      parts.pop();
      const next = parts.length ? parts.join('/') : '/';
      loadFiles(next);
    });
    back.appendChild(button);
    list.appendChild(back);
  }

  for (const entry of data.entries) {
    const li = document.createElement('li');
    const button = document.createElement('button');
    button.type = 'button';
    button.textContent = entry.Name;
    button.className = entry.IsDir ? 'entry-name entry-folder' : 'entry-name entry-file';
    button.addEventListener('click', () => {
      if (entry.IsDir) {
        loadFiles(entry.Path);
        return;
      }
      openFile(entry.Path);
    });

    const renameBtn = document.createElement('button');
    renameBtn.type = 'button';
    renameBtn.textContent = 'Rename';
    renameBtn.addEventListener('click', async (event) => {
      event.stopPropagation();
      const nextName = prompt('Rename to:', entry.Name);
      if (!nextName || nextName === entry.Name) return;
      const base = entry.Path.split('/').slice(0, -1).join('/');
      const target = base ? `${base}/${nextName}` : nextName;
      await apiJSON('/files/rename', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ oldPath: entry.Path, newPath: target }),
      });
      await loadFiles(window.currentListPath);
    });

    let downloadLink = null;
    if (!entry.IsDir) {
      downloadLink = document.createElement('a');
      downloadLink.className = 'file-action';
      downloadLink.textContent = 'Download';
      downloadLink.href = `/files/download?path=${encodeURIComponent(entry.Path)}`;
    }

    const deleteBtn = document.createElement('button');
    deleteBtn.type = 'button';
    deleteBtn.textContent = 'Delete';
    deleteBtn.addEventListener('click', async (event) => {
      event.stopPropagation();
      if (!confirm(`Delete ${entry.Name}?`)) return;
      await apiJSON('/files/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: entry.Path }),
      });
      await loadFiles(window.currentListPath);
    });

    li.appendChild(button);
    li.appendChild(renameBtn);
    if (downloadLink) li.appendChild(downloadLink);
    li.appendChild(deleteBtn);
    list.appendChild(li);
  }
}

async function openFile(path) {
  const data = await apiJSON(`/editor/open?path=${encodeURIComponent(path)}`);
  const currentPath = document.getElementById('current-path');
  const resolvedPath = data.Path || data.path || path;
  const content = data.Content || data.content || '';
  if (typeof window.setEditorDocument === 'function') {
    window.setEditorDocument(resolvedPath, content);
  } else {
    const editor = document.getElementById('editor');
    if (editor) editor.value = content;
  }
  if (currentPath) currentPath.textContent = `/${resolvedPath}`;
  const saveButton = document.getElementById('save-btn');
  if (saveButton) saveButton.disabled = false;
  window.currentEditorPath = resolvedPath;
  setWorkspaceView('split');
}

document.addEventListener('DOMContentLoaded', () => {
  setWorkspaceView('files');
  loadFiles('/').catch((error) => console.error(error));

  document.querySelectorAll('.view-button').forEach((button) => {
    button.addEventListener('click', () => setWorkspaceView(button.dataset.view));
  });

  const closeButton = document.getElementById('close-editor-btn');
  if (closeButton) closeButton.addEventListener('click', closeEditor);

  const folderBtn = document.getElementById('new-folder-btn');
  if (folderBtn) {
    folderBtn.addEventListener('click', async () => {
      const name = prompt('Folder name:');
      if (!name) return;
      const base = window.currentListPath === '/' ? '' : window.currentListPath;
      const target = base ? `${base}/${name}` : name;
      await apiJSON('/files/folder', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: target }),
      });
      await loadFiles(window.currentListPath);
    });
  }
});
