window.currentListPath = '/';
window.workspaceView = 'files';
window.selectedEntries = new Map();

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

function showToast(title, message, type = 'success') {
  const region = document.getElementById('toast-region');
  if (!region) return;
  const toast = document.createElement('div');
  toast.className = `app-toast${type === 'error' ? ' is-error' : ''}`;
  const content = document.createElement('div');
  const heading = document.createElement('strong');
  const detail = document.createElement('span');
  heading.textContent = title;
  detail.textContent = message;
  content.appendChild(heading);
  content.appendChild(detail);
  toast.appendChild(content);
  region.appendChild(toast);
  window.setTimeout(() => toast.remove(), 4000);
}

function closeInputDialog(value = null) {
  const dialog = document.getElementById('input-dialog');
  if (!dialog || !window.inputDialogResolve) return;
  const resolve = window.inputDialogResolve;
  window.inputDialogResolve = null;
  if (typeof dialog.close === 'function') {
    dialog.close();
  } else {
    dialog.removeAttribute('open');
  }
  resolve(value);
}

function requestInput({ title, label, value = '', action }) {
  const dialog = document.getElementById('input-dialog');
  const titleElement = document.getElementById('dialog-title');
  const labelElement = document.getElementById('dialog-label');
  const input = document.getElementById('dialog-input');
  const submit = document.getElementById('dialog-submit-btn');
  const error = document.getElementById('dialog-error');
  if (!dialog || !titleElement || !labelElement || !input || !submit || !error) {
    return Promise.resolve(null);
  }

  titleElement.textContent = title;
  labelElement.textContent = label;
  input.value = value;
  input.classList.remove('is-invalid');
  error.textContent = '';
  submit.textContent = action;
  if (typeof dialog.showModal === 'function') {
    dialog.showModal();
  } else {
    dialog.setAttribute('open', '');
  }
  window.setTimeout(() => {
    input.focus();
    input.select();
  }, 0);

  return new Promise((resolve) => {
    window.inputDialogResolve = resolve;
  });
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

function updateSelectionBar() {
  const bar = document.getElementById('selection-bar');
  const count = document.getElementById('selection-count');
  const unzipButton = document.getElementById('unzip-selected-btn');
  const entries = Array.from(window.selectedEntries.values());
  if (!bar || !count || !unzipButton) return;
  bar.hidden = entries.length === 0;
  count.textContent = `${entries.length} selected`;
  unzipButton.disabled = entries.length === 0 || entries.some((entry) => entry.IsDir || !entry.Name.toLowerCase().endsWith('.zip'));
}

function clearSelection() {
  window.selectedEntries.clear();
  document.querySelectorAll('.entry-select').forEach((checkbox) => {
    checkbox.checked = false;
  });
  document.querySelectorAll('.file-list li.is-selected').forEach((row) => {
    row.classList.remove('is-selected');
  });
  updateSelectionBar();
}

function selectedPaths() {
  return Array.from(window.selectedEntries.keys());
}

async function loadFiles(path = '/') {
  window.currentListPath = path;
  window.selectedEntries.clear();
  updateSelectionBar();
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
    const select = document.createElement('input');
    select.type = 'checkbox';
    select.className = 'entry-select';
    select.setAttribute('aria-label', `Select ${entry.Name}`);
    select.addEventListener('click', (event) => event.stopPropagation());
    select.addEventListener('change', () => {
      if (select.checked) {
        window.selectedEntries.set(entry.Path, entry);
      } else {
        window.selectedEntries.delete(entry.Path);
      }
      li.classList.toggle('is-selected', select.checked);
      updateSelectionBar();
    });
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
      const nextName = await requestInput({
        title: 'Rename item',
        label: `New name for ${entry.Name}`,
        value: entry.Name,
        action: 'Rename',
      });
      if (!nextName || nextName === entry.Name) return;
      const base = entry.Path.split('/').slice(0, -1).join('/');
      const target = base ? `${base}/${nextName}` : nextName;
      try {
        await apiJSON('/files/rename', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ oldPath: entry.Path, newPath: target }),
        });
        await loadFiles(window.currentListPath);
        showToast('Item renamed', `${entry.Name} is now ${nextName}.`);
      } catch (error) {
        showToast('Rename failed', error.message.trim(), 'error');
      }
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

    li.appendChild(select);
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

  const clearSelectionButton = document.getElementById('clear-selection-btn');
  if (clearSelectionButton) clearSelectionButton.addEventListener('click', clearSelection);

  const deleteSelectedButton = document.getElementById('delete-selected-btn');
  if (deleteSelectedButton) {
    deleteSelectedButton.addEventListener('click', async () => {
      const paths = selectedPaths();
      if (!paths.length || !confirm(`Delete ${paths.length} selected item(s)?`)) return;
      try {
        await apiJSON('/files/delete-many', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ paths }),
        });
        await loadFiles(window.currentListPath);
      } catch (error) {
        showToast('Delete failed', error.message.trim(), 'error');
      }
    });
  }

  const zipSelectedButton = document.getElementById('zip-selected-btn');
  if (zipSelectedButton) {
    zipSelectedButton.addEventListener('click', async () => {
      const paths = selectedPaths();
      if (!paths.length) return;
      let name = await requestInput({
        title: 'Create archive',
        label: 'Archive name',
        value: 'archive.zip',
        action: 'Create ZIP',
      });
      if (!name) return;
      if (!name.toLowerCase().endsWith('.zip')) name += '.zip';
      const base = window.currentListPath === '/' ? '' : window.currentListPath;
      const destination = base ? `${base}/${name}` : name;
      try {
        await apiJSON('/files/zip', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ paths, destination }),
        });
        await loadFiles(window.currentListPath);
      } catch (error) {
        showToast('ZIP failed', error.message.trim(), 'error');
      }
    });
  }

  const unzipSelectedButton = document.getElementById('unzip-selected-btn');
  if (unzipSelectedButton) {
    unzipSelectedButton.addEventListener('click', async () => {
      const paths = selectedPaths();
      if (!paths.length) return;
      try {
        await apiJSON('/files/unzip', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ paths }),
        });
        await loadFiles(window.currentListPath);
      } catch (error) {
        showToast('Unzip failed', error.message.trim(), 'error');
      }
    });
  }

  const fileBtn = document.getElementById('new-file-btn');
  if (fileBtn) {
    fileBtn.addEventListener('click', async () => {
      const name = await requestInput({
        title: 'Create file',
        label: 'File name',
        action: 'Create file',
      });
      if (!name) return;
      const base = window.currentListPath === '/' ? '' : window.currentListPath;
      const target = base ? `${base}/${name}` : name;
      try {
        await apiJSON('/files/file', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ path: target }),
        });
        await loadFiles(window.currentListPath);
        await openFile(target);
        showToast('File created', `${name} is ready to edit.`);
      } catch (error) {
        showToast('Create file failed', error.message.trim(), 'error');
      }
    });
  }

  const folderBtn = document.getElementById('new-folder-btn');
  if (folderBtn) {
    folderBtn.addEventListener('click', async () => {
      const name = await requestInput({
        title: 'Create folder',
        label: 'Folder name',
        action: 'Create folder',
      });
      if (!name) return;
      const base = window.currentListPath === '/' ? '' : window.currentListPath;
      const target = base ? `${base}/${name}` : name;
      try {
        await apiJSON('/files/folder', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ path: target }),
        });
        await loadFiles(window.currentListPath);
        showToast('Folder created', `${name} was added to this directory.`);
      } catch (error) {
        showToast('Create folder failed', error.message.trim(), 'error');
      }
    });
  }

  const dialog = document.getElementById('input-dialog');
  const dialogForm = document.getElementById('input-dialog-form');
  const dialogInput = document.getElementById('dialog-input');
  const dialogError = document.getElementById('dialog-error');
  if (dialogForm && dialogInput && dialogError) {
    dialogForm.addEventListener('submit', (event) => {
      event.preventDefault();
      const value = dialogInput.value.trim();
      if (!value) {
        dialogInput.classList.add('is-invalid');
        dialogError.textContent = 'A name is required.';
        dialogInput.focus();
        return;
      }
      closeInputDialog(value);
    });
    dialogInput.addEventListener('input', () => {
      dialogInput.classList.remove('is-invalid');
      dialogError.textContent = '';
    });
  }
  const cancelDialog = () => closeInputDialog(null);
  const dialogCancelButton = document.getElementById('dialog-cancel-btn');
  const dialogCloseButton = document.getElementById('dialog-close-btn');
  if (dialogCancelButton) dialogCancelButton.addEventListener('click', cancelDialog);
  if (dialogCloseButton) dialogCloseButton.addEventListener('click', cancelDialog);
  if (dialog) {
    dialog.addEventListener('cancel', (event) => {
      event.preventDefault();
      cancelDialog();
    });
  }
});
