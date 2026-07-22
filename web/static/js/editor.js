function modeForPath(path) {
  const extension = (path.split('.').pop() || '').toLowerCase();
  const modes = {
    js: 'javascript', jsx: 'javascript', ts: 'javascript', tsx: 'javascript', json: { name: 'javascript', json: true },
    css: 'css', html: 'htmlmixed', htm: 'htmlmixed', xml: 'xml', svg: 'xml',
    md: 'markdown', markdown: 'markdown', py: 'python', sh: 'shell', bash: 'shell', go: 'go',
  };
  return modes[extension] || null;
}

function getEditorValue() {
  if (window.codeEditor) return window.codeEditor.getValue();
  const textarea = document.getElementById('editor');
  return textarea ? textarea.value : '';
}

window.setEditorDocument = function setEditorDocument(path, content) {
  if (window.codeEditor) {
    window.codeEditor.setOption('mode', modeForPath(path));
    window.codeEditor.setValue(content);
    window.codeEditor.clearHistory();
    window.codeEditor.focus();
    return;
  }
  const textarea = document.getElementById('editor');
  if (textarea) textarea.value = content;
};

async function saveCurrentFile() {
  const path = window.currentEditorPath;
  if (!path) return;

  const response = await fetch('/editor/save', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ path, content: getEditorValue() }),
  });

  const status = document.getElementById('save-status');
  if (!response.ok) {
    if (status) {
      status.textContent = 'Save failed';
      status.className = 'save-status is-error';
    }
    return;
  }

  if (status) {
    status.textContent = 'Saved';
    status.className = 'save-status is-success';
    window.clearTimeout(window.saveStatusTimer);
    window.saveStatusTimer = window.setTimeout(() => {
      status.textContent = '';
      status.className = 'save-status';
    }, 2500);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  const textarea = document.getElementById('editor');
  if (!textarea) return;

  if (window.CodeMirror) {
    window.codeEditor = CodeMirror.fromTextArea(textarea, {
      lineNumbers: true,
      theme: 'material-darker',
      indentUnit: 2,
      tabSize: 2,
      lineWrapping: false,
      extraKeys: {
        'Ctrl-S': saveCurrentFile,
        'Cmd-S': saveCurrentFile,
      },
    });
  } else {
    textarea.addEventListener('keydown', (event) => {
      const isSave = (event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 's';
      if (!isSave) return;
      event.preventDefault();
      saveCurrentFile();
    });
  }

  const saveButton = document.getElementById('save-btn');
  if (saveButton) saveButton.addEventListener('click', saveCurrentFile);
});
