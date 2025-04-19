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

// Libraries array to store additional CQL libraries
let libraries = [];

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
  document.getElementById('librariesTabButton')
      .addEventListener('click', function(e) {
        showLibrariesTab();
      });
  document.getElementById('addLibrary')
      .addEventListener('click', function(e) {
        addLibrary();
      });
  document.getElementById('uploadLibrary')
      .addEventListener('change', handleFileUpload);
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
    }
  };
  xhr.open('POST', '/eval_cql', true);
  xhr.setRequestHeader('Content-Type', 'text/json');
  
  // Collect library content to send with request
  const libraryContents = libraries.map(lib => lib.content);
  
  xhr.send(JSON.stringify({
    'cql': code, 
    'data': data,
    'libraries': libraryContents
  }));
}

/**
 * showDataTab shows the data tab and hides the CQL tab.
 */
function showDataTab() {
  document.getElementById('cqlEntry').style.display = 'none';
  document.getElementById('dataEntry').style.display = 'block';
  document.getElementById('librariesEntry').style.display = 'none';

  document.getElementById('dataTabButton').className = 'active';
  document.getElementById('cqlTabButton').className = '';
  document.getElementById('librariesTabButton').className = '';
}

/**
 * showCQLTab shows the CQL tab and hides the data tab.
 */
function showCQLTab() {
  document.getElementById('cqlEntry').style.display = 'block';
  document.getElementById('dataEntry').style.display = 'none';
  document.getElementById('librariesEntry').style.display = 'none';

  document.getElementById('cqlTabButton').className = 'active';
  document.getElementById('dataTabButton').className = '';
  document.getElementById('librariesTabButton').className = '';
}

/**
 * showLibrariesTab shows the Libraries tab and hides other tabs.
 */
function showLibrariesTab() {
  document.getElementById('cqlEntry').style.display = 'none';
  document.getElementById('dataEntry').style.display = 'none';
  document.getElementById('librariesEntry').style.display = 'block';

  document.getElementById('librariesTabButton').className = 'active';
  document.getElementById('cqlTabButton').className = '';
  document.getElementById('dataTabButton').className = '';
}

/**
 * addLibrary adds a new library to the libraries list and updates the UI.
 */
function addLibrary(name = '', content = '') {
  const libraryId = Date.now(); // Unique ID for the library
  
  // Add to libraries array
  libraries.push({
    id: libraryId,
    name: name,
    content: content
  });
  
  // Update the UI
  renderLibraries();
  
  // Save to localStorage
  saveLibrariesToLocalStorage();
}

/**
 * removeLibrary removes a library from the libraries list and updates the UI.
 */
function removeLibrary(libraryId) {
  libraries = libraries.filter(lib => lib.id !== libraryId);
  renderLibraries();
  saveLibrariesToLocalStorage();
}

/**
 * renderLibraries updates the libraries UI with the current libraries.
 */
function renderLibraries() {
  const container = document.getElementById('librariesContainer');
  container.innerHTML = '';
  
  libraries.forEach(library => {
    const libraryContainer = document.createElement('div');
    libraryContainer.className = 'libraryContainer';
    
    const headerDiv = document.createElement('div');
    headerDiv.className = 'libraryHeader';
    
    // Create label for the library name input
    const nameLabel = document.createElement('div');
    nameLabel.className = 'libraryNameLabel';
    
    const nameInput = document.createElement('input');
    nameInput.type = 'text';
    nameInput.value = library.name;
    nameInput.placeholder = 'Library Name';
    nameInput.className = 'libraryNameInput';
    nameInput.oninput = function(e) {
      library.name = e.target.value;
      saveLibrariesToLocalStorage();
    };
    
    nameLabel.appendChild(document.createTextNode('Library Name:'));
    nameLabel.appendChild(nameInput);
    
    const removeButton = document.createElement('button');
    removeButton.className = 'removeLibraryButton';
    removeButton.textContent = 'Remove';
    removeButton.onclick = function() {
      removeLibrary(library.id);
    };
    
    headerDiv.appendChild(nameLabel);
    headerDiv.appendChild(removeButton);
    
    const editorDiv = document.createElement('div');
    editorDiv.className = 'codeInputContainer';
    
    const codeInput = document.createElement('code-input');
    codeInput.setAttribute('lang', 'cql');
    codeInput.setAttribute('placeholder', 'Type CQL Library Here');
    codeInput.className = 'codeInput';
    codeInput.value = library.content;
    codeInput.onchange = function(e) {
      library.content = e.target.value;
      saveLibrariesToLocalStorage();
    };
    
    editorDiv.appendChild(codeInput);
    
    libraryContainer.appendChild(headerDiv);
    libraryContainer.appendChild(editorDiv);
    
    container.appendChild(libraryContainer);
  });
}

/**
 * saveLibrariesToLocalStorage saves the libraries to localStorage.
 */
function saveLibrariesToLocalStorage() {
  localStorage.setItem('cqlLibraries', JSON.stringify(libraries));
}

/**
 * loadLibrariesFromLocalStorage loads the libraries from localStorage.
 */
function loadLibrariesFromLocalStorage() {
  const storedLibraries = localStorage.getItem('cqlLibraries');
  if (storedLibraries) {
    libraries = JSON.parse(storedLibraries);
    renderLibraries();
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
 * handleFileUpload processes uploaded CQL library files
 */
function handleFileUpload(event) {
  const fileList = event.target.files;
  if (fileList.length === 0) {
    return; // No file selected
  }
  
  const file = fileList[0];
  const reader = new FileReader();
  
  reader.onload = function(e) {
    const content = e.target.result;
    // Extract library name from filename (remove .cql extension)
    const fileName = file.name.replace(/\.cql$/i, '');
    
    // Add the library with the file content
    addLibrary(fileName, content);
  };
  
  reader.readAsText(file);
  
  // Reset the file input so the same file can be selected again
  event.target.value = '';
}

/**
 * main is the entrypoint for the script.
 */
function main() {
  setupPrism();
  updateInputs();
  bindInputsOnChange();
  bindButtonActions();

  // Load libraries from localStorage
  loadLibrariesFromLocalStorage();

  // Initially hide non-CQL tabs
  document.getElementById('dataEntry').style.display = 'none';
  document.getElementById('librariesEntry').style.display = 'none';
  
  // Set CQL tab as active
  document.getElementById('cqlTabButton').className = 'active';
}

main();  // All code actually executed when the script is loaded by the HTML.