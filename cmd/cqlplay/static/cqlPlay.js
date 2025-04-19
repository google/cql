/**
 * Copyright 2024 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * @fileoverview cqlPlay.js contains all Javascript needed for the cql
 * playground frontend.
 *
 * The goal is to be as simple as possible. This is meant to be a non-production
 * playground entrypoint to our CQL engine. This is not meant to be run on a
 * production server, rather it's meant to be run locally as a way to play with
 * the engine.
 */

import {syntheticPatient} from './syntheticPatient.js';

// These are globals that are bound to the onchange event of the code inputs
// and syntax highlighted outputs. They are seeded with initial data for the
// inputs.

let code = `library Explore version '1.2.3'
using FHIR version '4.0.1'

include FHIRHelpers version '4.0.1' called FHIRHelpers

context Patient`;

let data = syntheticPatient;

let results = '';

// Helper functions:

/**
 * updateInputs updates the inputs to the code and data input boxes to match
 * the code and data global variables.
 */
function updateInputs() {
  document.getElementById('cqlInput').value = code;
  document.getElementById('dataInput').value = data;
}

/**
 * bindInputsOnChange binds the onchange events of the code and data input boxes
 * to the code and data global variables.
 */
function bindInputsOnChange() {
  document.getElementById('cqlInput').onchange = function(e) {
    code = e.target.value;
  };
  document.getElementById('dataInput').onchange = function(e) {
    data = e.target.value;
  };
}

/**
 * bindButtonActions binds the actions of the buttons to their respective
 * functions.
 */
function bindButtonActions() {
  document.getElementById('submit').addEventListener('click', function(e) {
    runCQL();
  });
  document.getElementById('cqlTabButton')
      .addEventListener('click', function(e) {
        showCQLTab();
      });
  document.getElementById('dataTabButton')
      .addEventListener('click', function(e) {
        showDataTab();
      });
  document.getElementById('filesTabButton')
      .addEventListener('click', function(e) {
        showFilesTab();
      });
  
  // File management buttons
  document.getElementById('browseButton').addEventListener('click', function(e) {
    document.getElementById('fileInput').click();
  });
  document.getElementById('fileInput').addEventListener('change', function(e) {
    handleFileSelect(e);
  });
  
  // Set up drag and drop events
  const dropZone = document.getElementById('dropZone');
  dropZone.addEventListener('dragover', handleDragOver);
  dropZone.addEventListener('dragleave', handleDragLeave);
  dropZone.addEventListener('drop', handleDrop);
}

/**
 * runCQL runs the CQL code in the code input box and displays the results in
 * the results box.
 */
function runCQL() {
  let xhr = new XMLHttpRequest();
  xhr.onreadystatechange = function() {
    if (xhr.readyState == XMLHttpRequest.DONE) {
      document.getElementById('results').innerHTML = xhr.responseText;
      Prism.highlightAll();
      results = xhr.responseText;
      
      // Save state after running CQL
      saveState();
    }
  };
  xhr.open('POST', '/eval_cql', true);
  xhr.setRequestHeader('Content-Type', 'text/json');
  xhr.send(JSON.stringify({'cql': code, 'data': data}));
}

/**
 * showDataTab shows the data tab and hides the other tabs.
 */
function showDataTab() {
  document.getElementById('cqlEntry').style.display = 'none';
  document.getElementById('dataEntry').style.display = 'block';
  document.getElementById('filesEntry').style.display = 'none';

  document.getElementById('dataTabButton').className = 'active';
  document.getElementById('cqlTabButton').className = '';
  document.getElementById('filesTabButton').className = '';
  
  // Save state when tab is changed
  saveState();
}

/**
 * showCQLTab shows the CQL tab and hides the other tabs.
 */
function showCQLTab() {
  document.getElementById('cqlEntry').style.display = 'block';
  document.getElementById('dataEntry').style.display = 'none';
  document.getElementById('filesEntry').style.display = 'none';

  document.getElementById('cqlTabButton').className = 'active';
  document.getElementById('dataTabButton').className = '';
  document.getElementById('filesTabButton').className = '';
  
  // Save state when tab is changed
  saveState();
}

/**
 * showFilesTab shows the Files tab and hides the other tabs.
 */
function showFilesTab() {
  document.getElementById('cqlEntry').style.display = 'none';
  document.getElementById('dataEntry').style.display = 'none';
  document.getElementById('filesEntry').style.display = 'block';

  document.getElementById('filesTabButton').className = 'active';
  document.getElementById('cqlTabButton').className = '';
  document.getElementById('dataTabButton').className = '';
  
  // Refresh the file list when showing the tab
  listFiles();
  
  // Save state when tab is changed
  saveState();
}

/**
 * handleDragOver handles the dragover event for the drop zone.
 */
function handleDragOver(e) {
  e.preventDefault();
  e.stopPropagation();
  this.classList.add('dragover');
}

/**
 * handleDragLeave handles the dragleave event for the drop zone.
 */
function handleDragLeave(e) {
  e.preventDefault();
  e.stopPropagation();
  this.classList.remove('dragover');
}

/**
 * handleDrop handles the drop event for the drop zone.
 */
function handleDrop(e) {
  e.preventDefault();
  e.stopPropagation();
  this.classList.remove('dragover');
  
  const files = e.dataTransfer.files;
  if (files.length > 0) {
    uploadFiles(files);
  }
}

/**
 * handleFileSelect handles the file selection from the file input.
 */
function handleFileSelect(e) {
  const files = e.target.files;
  if (files.length > 0) {
    uploadFiles(files);
  }
}

/**
 * uploadFiles uploads the selected files to the server.
 */
function uploadFiles(files) {
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    
    // Only upload .cql files
    if (!file.name.toLowerCase().endsWith('.cql')) {
      alert('Only .cql files are supported');
      continue;
    }
    
    // Create a FormData object
    const formData = new FormData();
    formData.append('file', file);
    
    // Create a new XMLHttpRequest
    const xhr = new XMLHttpRequest();
    
    // Add the file to the UI with progress indicator
    const fileId = 'file-' + Date.now() + '-' + i;
    addFileToUI(fileId, file.name, 'Uploading...', true);
    
    // Set up the request
    xhr.open('POST', '/upload_file', true);
    
    // Set up event handlers
    xhr.onload = function() {
      if (xhr.status === 200) {
        // Update the file in the UI
        updateFileInUI(fileId, file.name, JSON.parse(xhr.responseText).size, false);
      } else {
        // Remove the file from the UI
        removeFileFromUI(fileId);
        alert('Upload failed: ' + xhr.responseText);
      }
    };
    
    xhr.onerror = function() {
      // Remove the file from the UI
      removeFileFromUI(fileId);
      alert('Upload failed');
    };
    
    xhr.upload.onprogress = function(e) {
      if (e.lengthComputable) {
        const percentComplete = (e.loaded / e.total) * 100;
        updateProgressInUI(fileId, percentComplete);
      }
    };
    
    // Send the request
    xhr.send(formData);
  }
}

/**
 * addFileToUI adds a file to the UI.
 */
function addFileToUI(id, name, size, uploading) {
  // Remove the "No files" message if it exists
  const noFilesMessage = document.getElementById('noFilesMessage');
  if (noFilesMessage) {
    noFilesMessage.remove();
  }
  
  // Create the file item
  const fileItem = document.createElement('div');
  fileItem.id = id;
  fileItem.className = 'fileItem';
  
  // Create the file info
  const fileInfo = document.createElement('div');
  fileInfo.className = 'fileInfo';
  
  // Create the file name
  const fileName = document.createElement('div');
  fileName.className = 'fileName';
  fileName.textContent = name;
  fileInfo.appendChild(fileName);
  
  // Create the file size
  const fileSize = document.createElement('div');
  fileSize.className = 'fileSize';
  fileSize.textContent = size;
  fileInfo.appendChild(fileSize);
  
  // Add the file info to the file item
  fileItem.appendChild(fileInfo);
  
  // If uploading, add a progress bar
  if (uploading) {
    const uploadProgress = document.createElement('div');
    uploadProgress.className = 'uploadProgress';
    
    const uploadProgressBar = document.createElement('div');
    uploadProgressBar.className = 'uploadProgressBar';
    uploadProgressBar.style.width = '0%';
    
    uploadProgress.appendChild(uploadProgressBar);
    fileInfo.appendChild(uploadProgress);
  } else {
    // Create the file actions
    const fileActions = document.createElement('div');
    fileActions.className = 'fileActions';
    
    // Create the delete button
    const deleteButton = document.createElement('button');
    deleteButton.className = 'deleteButton';
    deleteButton.textContent = 'Delete';
    deleteButton.addEventListener('click', function() {
      deleteFile(name, id);
    });
    
    fileActions.appendChild(deleteButton);
    fileItem.appendChild(fileActions);
  }
  
  // Add the file item to the uploaded files container
  document.getElementById('uploadedFiles').appendChild(fileItem);
}

/**
 * updateFileInUI updates a file in the UI.
 */
function updateFileInUI(id, name, size, uploading) {
  const fileItem = document.getElementById(id);
  if (!fileItem) return;
  
  // Update the file size
  const fileSize = fileItem.querySelector('.fileSize');
  fileSize.textContent = size;
  
  // Remove the progress bar if it exists
  const uploadProgress = fileItem.querySelector('.uploadProgress');
  if (uploadProgress) {
    uploadProgress.remove();
  }
  
  // If not uploading, add the delete button
  if (!uploading && !fileItem.querySelector('.fileActions')) {
    // Create the file actions
    const fileActions = document.createElement('div');
    fileActions.className = 'fileActions';
    
    // Create the delete button
    const deleteButton = document.createElement('button');
    deleteButton.className = 'deleteButton';
    deleteButton.textContent = 'Delete';
    deleteButton.addEventListener('click', function() {
      deleteFile(name, id);
    });
    
    fileActions.appendChild(deleteButton);
    fileItem.appendChild(fileActions);
  }
}

/**
 * updateProgressInUI updates the progress bar in the UI.
 */
function updateProgressInUI(id, percent) {
  const fileItem = document.getElementById(id);
  if (!fileItem) return;
  
  const uploadProgressBar = fileItem.querySelector('.uploadProgressBar');
  if (uploadProgressBar) {
    uploadProgressBar.style.width = percent + '%';
  }
}

/**
 * removeFileFromUI removes a file from the UI.
 */
function removeFileFromUI(id) {
  const fileItem = document.getElementById(id);
  if (fileItem) {
    fileItem.remove();
  }
  
  // If there are no more files, add the "No files" message
  const uploadedFiles = document.getElementById('uploadedFiles');
  if (uploadedFiles.children.length === 0) {
    const noFilesMessage = document.createElement('p');
    noFilesMessage.id = 'noFilesMessage';
    noFilesMessage.textContent = 'No files uploaded yet';
    uploadedFiles.appendChild(noFilesMessage);
  }
}

/**
 * deleteFile deletes a file from the server.
 */
function deleteFile(filename, id) {
  if (!confirm('Are you sure you want to delete ' + filename + '?')) {
    return;
  }
  
  const xhr = new XMLHttpRequest();
  xhr.open('POST', '/delete_file', true);
  xhr.setRequestHeader('Content-Type', 'application/json');
  
  xhr.onload = function() {
    if (xhr.status === 200) {
      // Remove the file from the UI
      removeFileFromUI(id);
    } else {
      alert('Delete failed: ' + xhr.responseText);
    }
  };
  
  xhr.onerror = function() {
    alert('Delete failed');
  };
  
  xhr.send(JSON.stringify({ filename: filename }));
}

/**
 * listFiles lists the files on the server.
 */
function listFiles() {
  const xhr = new XMLHttpRequest();
  xhr.open('GET', '/list_files', true);
  
  xhr.onload = function() {
    if (xhr.status === 200) {
      // Clear the uploaded files container
      const uploadedFiles = document.getElementById('uploadedFiles');
      uploadedFiles.innerHTML = '';
      
      // Parse the response
      const response = JSON.parse(xhr.responseText);
      
      // If there are no files, add the "No files" message
      if (response.files.length === 0) {
        const noFilesMessage = document.createElement('p');
        noFilesMessage.id = 'noFilesMessage';
        noFilesMessage.textContent = 'No files uploaded yet';
        uploadedFiles.appendChild(noFilesMessage);
        return;
      }
      
      // Add each file to the UI
      for (let i = 0; i < response.files.length; i++) {
        const file = response.files[i];
        const fileId = 'file-' + i;
        addFileToUI(fileId, file.name, formatFileSize(file.size), false);
      }
    } else {
      alert('Failed to list files: ' + xhr.responseText);
    }
  };
  
  xhr.onerror = function() {
    alert('Failed to list files');
  };
  
  xhr.send();
}

/**
 * formatFileSize formats a file size in bytes to a human-readable string.
 */
function formatFileSize(bytes) {
  if (bytes < 1024) {
    return bytes + ' bytes';
  } else if (bytes < 1024 * 1024) {
    return (bytes / 1024).toFixed(2) + ' KB';
  } else {
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
  }
}

/**
 * setupPrism sets up the Prism syntax highlighting library.
 */
function setupPrism() {
  // Inspired by
  // https://github.com/cqframework/cqf/blob/f0aa9ade146aa03cc1cc7732583ae97385310cf6/input/images/cql.js
  // which is Apache 2.0 Licensed:
  // https://github.com/cqframework/cqf/blob/f0aa9ade146aa03cc1cc7732583ae97385310cf6/LICENSE
  Prism.languages.cql = {
    'comment': {
      pattern: /(^|[^\\])(?:\/\*[\s\S]*?\*\/|(?:\/\/|#).*)/,
      lookbehind: true
    },
    'string': {pattern: /(')(?:\\[\s\S]|(?!\1)[^\\]|\1\1)*\1/, greedy: true},
    'variable': {pattern: /(["`])(?:\\[\s\S]|(?!\1)[^\\])+\1/, greedy: true},
    'keyword':
        /\b(?:after|all|and|as|asc|ascending|before|between|by|called|case|cast|code|Code|codesystem|codesystems|collapse|concept|Concept|contains|context|convert|date|day|days|default|define|desc|descending|difference|display|distinct|div|duration|during|else|end|ends|except|exists|expand|false|flatten|from|function|hour|hours|if|implies|in|include|includes|included in|intersect|Interval|is|let|library|List|maximum|meets|millisecond|milliseconds|minimum|minute|minutes|mod|month|months|not|null|occurs|of|on|or|overlaps|parameter|per|predecessor|private|properly|public|return|same|singleton|second|seconds|start|starts|sort|successor|such that|then|time|timezoneoffset|to|true|Tuple|union|using|valueset|version|week|weeks|where|when|width|with|within|without|xor|year|years)\b/i,
    'boolean': /\b(?:null|false|null)\b/i,
    'number': /\b0x[\da-f]+\b|\b\d+(?:\.\d*)?|\B\.\d+\b/i,
    'punctuation': /[;[\]()`,.]/,
    'operator': /[-+*\/=%^~]|&&?|\|\|?|!=?|<(?:=>?|<|>)?|>[>=]?\b/i
  };

  codeInput.registerTemplate(
      'syntax-highlighted', codeInput.templates.prism(Prism, []));
}

/**
 * saveState saves the current state to localStorage.
 */
function saveState() {
  localStorage.setItem('cqlplay_code', code);
  localStorage.setItem('cqlplay_data', data);
  localStorage.setItem('cqlplay_results', results);
  
  // Save active tab
  let activeTab = 'cql';
  if (document.getElementById('dataTabButton').className === 'active') {
    activeTab = 'data';
  } else if (document.getElementById('filesTabButton').className === 'active') {
    activeTab = 'files';
  }
  localStorage.setItem('cqlplay_activeTab', activeTab);
}

/**
 * restoreState restores the state from localStorage.
 */
function restoreState() {
  if (localStorage.getItem('cqlplay_code')) {
    code = localStorage.getItem('cqlplay_code');
    document.getElementById('cqlInput').value = code;
  }
  
  if (localStorage.getItem('cqlplay_data')) {
    data = localStorage.getItem('cqlplay_data');
    document.getElementById('dataInput').value = data;
  }
  
  if (localStorage.getItem('cqlplay_results')) {
    results = localStorage.getItem('cqlplay_results');
    document.getElementById('results').innerHTML = results;
    Prism.highlightAll();
  }
  
  // Restore active tab
  if (localStorage.getItem('cqlplay_activeTab')) {
    const activeTab = localStorage.getItem('cqlplay_activeTab');
    if (activeTab === 'data') {
      showDataTab();
    } else if (activeTab === 'files') {
      showFilesTab();
    } else {
      showCQLTab();
    }
  }
}

/**
 * startHealthCheck starts a health check interval to detect server restarts.
 */
function startHealthCheck() {
  let serverWasDown = false;
  
  // Check server health every 2 seconds
  setInterval(() => {
    fetch('/health', { method: 'GET' })
      .then(response => {
        if (response.ok && serverWasDown) {
          // Server was down but is now up - refresh
          console.log('Server restarted, refreshing page...');
          serverWasDown = false;
          // Save state before refreshing
          saveState();
          window.location.reload();
        } else if (response.ok) {
          // Server is up and was not down before
          serverWasDown = false;
        }
      })
      .catch(() => {
        // Server is down
        if (!serverWasDown) {
          console.log('Server is down, waiting for it to come back up...');
          // Save state when server first goes down
          saveState();
        }
        serverWasDown = true;
      });
  }, 2000);
}

/**
 * main is the entrypoint for the script.
 */
function main() {
  setupPrism();
  updateInputs();
  bindInputsOnChange();
  bindButtonActions();

  // Initially hide dataEntry and filesEntry tabs:
  document.getElementById('dataEntry').style.display = 'none';
  document.getElementById('filesEntry').style.display = 'none';
  
  // Restore state from localStorage
  restoreState();
  
  // Start health check for hot reload
  startHealthCheck();
  
  // Add input event listeners for state saving
  document.getElementById('cqlInput').addEventListener('input', function(e) {
    code = e.target.value;
    saveState();
  });
  
  document.getElementById('dataInput').addEventListener('input', function(e) {
    data = e.target.value;
    saveState();
  });
}

main();  // All code actually executed when the script is loaded by the HTML.
