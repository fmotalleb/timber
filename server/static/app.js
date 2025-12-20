let AUTH_HEADER = null;
let tailInterval = null;

// --- DOM Elements ---
const filesContainer = document.getElementById("files");
const output = document.getElementById("output");
const jsonViewer = document.getElementById("json-viewer");
const jsonOutput = document.getElementById("json-output");

// --- Core API & Utility Functions ---

function encodePath(path) {
    return encodeURIComponent(path);
}

function stopFollow() {
    if (tailInterval) {
        clearInterval(tailInterval);
        tailInterval = null;
    }
}

async function authFetch(url) {
    const res = await fetch(url, {
        headers: {
            "Authorization": AUTH_HEADER
        }
    });

    if (!res.ok) {
        throw new Error(res.status + " " + res.statusText);
    }

    return res;
}

// --- Content Rendering ---

function renderOutput(text) {
    output.innerHTML = ""; // Clear previous output
    const lines = text.split("\n");
    lines.forEach(line => {
        const lineEl = document.createElement("div");
        lineEl.className = "log-line";

        const logLevels = ["WARN", "FINE", "ERROR", "OK", "INFO"];
        let hasLevel = false;
        logLevels.forEach(level => {
            if (line.includes(level)) {
                const regex = new RegExp(`(${level}[^\s\\]\n]*)`, 'g');
                lineEl.innerHTML = line.replace(regex, `<span class=" ${level}">$1</span>`);
                hasLevel = true;
            }
        });

        if (!hasLevel) {
            lineEl.textContent = line;
        }

        output.appendChild(lineEl);
    });
    output.scrollTop = output.scrollHeight;
}

async function fetchText(url) {
    const res = await authFetch(url);
    const text = await res.text();
    // Use the wrapped renderOutput to ensure search state is cleared
    wrappedRenderOutput(text);
}

// --- Authentication ---

async function login() {
    const user = document.getElementById("user").value;
    const pass = document.getElementById("pass").value;

    AUTH_HEADER = "Basic " + btoa(user + ":" + pass);

    try {
        await authFetch("./filesystem/ls");

        localStorage.setItem("auth_user", user);
        localStorage.setItem("auth_pass", pass);

        document.getElementById("login").classList.add("hidden");
        document.getElementById("app").classList.remove("hidden");

        loadFiles();
    } catch (e) {
        document.getElementById("login-error").textContent = "Authentication failed";
        AUTH_HEADER = null;
    }
}

function logout() {
    localStorage.removeItem("auth_user");
    localStorage.removeItem("auth_pass");
    AUTH_HEADER = null;
    document.getElementById("login").classList.remove("hidden");
    document.getElementById("app").classList.add("hidden");
    filesContainer.innerHTML = "";
    output.innerHTML = "";
}

// --- File Tree Navigation ---

function filterFiles() {
    const filterText = document.getElementById("file-filter").value.toLowerCase();

    function recursiveFilter(node) {
        const nodeNameEl = node.querySelector(":scope > .node-name");
        if (!nodeNameEl) return false;
        
        const nodeName = nodeNameEl.textContent.toLowerCase();
        const selfMatches = filterText === '' || nodeName.includes(filterText);

        const childrenContainer = node.querySelector(":scope > .node-children");
        let hasVisibleChild = false;
        if (childrenContainer) {
            for (const childNode of childrenContainer.children) {
                if (recursiveFilter(childNode)) {
                    hasVisibleChild = true;
                }
            }
        }

        const isVisible = selfMatches || hasVisibleChild;
        node.style.display = isVisible ? "block" : "none";

        if (hasVisibleChild && !selfMatches && filterText) {
            childrenContainer.classList.add("open");
            nodeNameEl.classList.add("open");
        }
        
        return isVisible;
    }

    for (const node of filesContainer.children) {
        recursiveFilter(node);
    }
}

function createNode(node) {
    // Compact single-child directories like GitHub
    if (node.type === "dir" && node.children && node.children.length === 1 && node.children[0].type === "dir") {
        const child = node.children[0];
        const mergedNode = {
            ...child,
            name: node.name + "/" + child.name,
        };
        // Continue compacting by recursively calling createNode
        return createNode(mergedNode);
    }

    const nodeEl = document.createElement("div");
    nodeEl.className = "tree-node";
    nodeEl.dataset.type = node.type;

    const nodeName = document.createElement("div");
    nodeName.className = "node-name";
    nodeName.textContent = node.name;

    if (node.type === "dir") {
        const childrenEl = document.createElement("div");
        childrenEl.className = "node-children";
        if (node.children) {
            node.children.forEach(child => childrenEl.appendChild(createNode(child)));
        }
        nodeEl.appendChild(nodeName);
        nodeEl.appendChild(childrenEl);

        nodeName.addEventListener("click", (e) => {
            // Stop propagation to prevent file controls from triggering this
            e.stopPropagation();
            childrenEl.classList.toggle("open");
            nodeName.classList.toggle("open");
        });
    } else { // It's a file
        nodeEl.className += " file";
        const path = node.path;
        
        nodeEl.appendChild(nodeName);

        const controls = document.createElement('div');
        controls.className = 'file-controls';
        controls.style.display = 'flex';
        controls.style.gap = '8px';
        controls.style.alignItems = 'center';

        const lines = document.createElement("input");
        lines.type = "number";
        lines.min = 1;
        lines.value = 10;
        lines.style.width = "50px";

        const cat = document.createElement("button");
        cat.textContent = "cat";
        cat.onclick = () => { stopFollow(); fetchText(`./filesystem/cat?path=${encodePath(path)}`); };

        const head = document.createElement("button");
        head.textContent = "head";
        head.onclick = () => { stopFollow(); fetchText(`./filesystem/head?path=${encodePath(path)}&lines=${lines.value}`); };

        const tail = document.createElement("button");
        tail.textContent = "tail";
        tail.onclick = () => { stopFollow(); fetchText(`./filesystem/tail?path=${encodePath(path)}&lines=${lines.value}&follow=false`); };
        
        const follow = document.createElement("button");
        follow.textContent = "follow";
        follow.onclick = async () => {
            stopFollow();
            wrappedRenderOutput(""); // Clear output
            const url = `./filesystem/tail?path=${encodePath(path)}&lines=${lines.value}&follow=true`;
            
            try {
                const res = await authFetch(url);
                const reader = res.body.getReader();
                const decoder = new TextDecoder();
    
                async function readChunk() {
                    try {
                        const { done, value } = await reader.read();
                        if (done) return;
                        
                        const chunk = decoder.decode(value, { stream: true });
                        const lines = chunk.split("\n");
                        
                        lines.forEach(line => {
                            if (line.trim() === "") return;
                            const lineEl = document.createElement("div");
                            lineEl.className = "log-line";
                            lineEl.textContent = line; // Simplified rendering for streaming
                            output.appendChild(lineEl);
                        });
                        
                        output.scrollTop = output.scrollHeight;
                        readChunk();
                    } catch (e) {
                        // Handle stream reading errors
                        console.error("Stream reading error:", e);
                    }
                }
                readChunk();
            } catch (e) {
                wrappedRenderOutput(`\n[Error: ${e.message}]`);
            }
        };

        const download = document.createElement("button");
        download.textContent = "download";
        download.onclick = async () => {
            stopFollow();
            try {
                const res = await authFetch(`./filesystem/cat?path=${encodePath(path)}`);
                const blob = await res.blob();
                const url = URL.createObjectURL(blob);
                const a = document.createElement("a");
                a.href = url;
                a.download = path.split("/").pop();
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                URL.revokeObjectURL(url);
            } catch (e) {
                wrappedRenderOutput(`\n[Error: ${e.message}]`);
            }
        };
        
        const viewJson = document.createElement("button");
        viewJson.textContent = "View as JSON";
        viewJson.onclick = async () => {
            stopFollow();
            try {
                const res = await authFetch(`./filesystem/cat?path=${encodePath(path)}`);
                const text = await res.text();
                openJsonViewer(text);
            } catch (e) {
                wrappedRenderOutput(`Error fetching file for JSON view.\n\n[Error: ${e.message}]`);
            }
        };

        controls.append(lines, cat, head, tail, follow, download, viewJson);
        nodeEl.appendChild(controls);
    }
    return nodeEl;
}

async function loadFiles() {
    const res = await authFetch("./filesystem/ls");
    const tree = await res.json();

    filesContainer.innerHTML = "";
    if (tree) {
        tree.forEach(node => filesContainer.appendChild(createNode(node)));
    }
}

// --- JSON Viewer ---

function openJsonViewer(fileContent) {
    jsonOutput.innerHTML = "";
    const lines = fileContent.split('\n').filter(line => line.trim() !== "");
    const jsonObjects = [];

    lines.forEach(line => {
        try {
            jsonObjects.push(JSON.parse(line));
        } catch (e) { /* Not a valid JSON line, ignore it. */ }
    });

    if (jsonObjects.length === 0) {
        jsonOutput.textContent = "No valid JSON objects found per line in this file.";
        jsonViewer.classList.remove("hidden");
        return;
    }

    const table = document.createElement('table');
    table.className = 'json-table';

    const thead = document.createElement('thead');
    const headerRow = document.createElement('tr');
    const headers = [...new Set(jsonObjects.flatMap(obj => Object.keys(obj)))];
    headers.forEach(headerText => {
        const th = document.createElement('th');
        th.textContent = headerText;
        headerRow.appendChild(th);
    });
    thead.appendChild(headerRow);
    table.appendChild(thead);

    const tbody = document.createElement('tbody');
    jsonObjects.forEach(obj => {
        const row = document.createElement('tr');
        headers.forEach(header => {
            const cell = document.createElement('td');
            const value = obj[header];
            cell.textContent = (typeof value === 'object' && value !== null) ? JSON.stringify(value, null, 2) : value;
            row.appendChild(cell);
        });
        tbody.appendChild(row);
    });
    table.appendChild(tbody);

    jsonOutput.appendChild(table);
    jsonViewer.classList.remove("hidden");
}

function closeJsonViewer() {
    jsonViewer.classList.add("hidden");
}

// --- In-File Search ---

let searchMatches = [];
let currentSearchIndex = -1;
let debounceTimer;
let originalOutputHTML = "";
let isSearching = false;

function performSearch() {
    const searchTerm = document.getElementById("content-search").value;

    if (searchTerm.length < 2) {
        if (isSearching) {
            output.innerHTML = originalOutputHTML;
            isSearching = false;
            originalOutputHTML = "";
        }
        searchMatches = [];
        currentSearchIndex = -1;
        return;
    }

    if (!isSearching) {
        originalOutputHTML = output.innerHTML;
        isSearching = true;
    }
    
    output.innerHTML = originalOutputHTML;

    const regex = new RegExp(`(${escapeRegExp(searchTerm)})`, 'gi');
    output.innerHTML = output.innerHTML.replace(regex, '<span class="highlight">$1</span>');

    searchMatches = output.getElementsByClassName("highlight");
    if (searchMatches.length > 0) {
        currentSearchIndex = 0;
        highlightCurrentMatch();
    } else {
        currentSearchIndex = -1;
    }
}

function navigateSearch(direction) {
    if (searchMatches.length === 0) return;
    currentSearchIndex += direction;
    if (currentSearchIndex < 0) currentSearchIndex = searchMatches.length - 1;
    if (currentSearchIndex >= searchMatches.length) currentSearchIndex = 0;
    highlightCurrentMatch();
}

function highlightCurrentMatch() {
    searchMatches.forEach(match => match.classList.remove("highlight-current"));
    const currentMatch = searchMatches[currentSearchIndex];
    if (currentMatch) {
        currentMatch.classList.add("highlight-current");
        currentMatch.scrollIntoView({ behavior: "smooth", block: "center" });
    }
}

function escapeRegExp(string) {
    return string.replace(/[.*+?^${}()|[\\]/g, '\\$&');
}

// Wrap renderOutput to reset search state whenever new content is loaded
const wrappedRenderOutput = (text) => {
    document.getElementById("content-search").value = "";
    searchMatches = [];
    currentSearchIndex = -1;
    isSearching = false;
    originalOutputHTML = "";
    renderOutput(text);
};


// --- App Initialization ---

window.addEventListener("load", async () => {
    // Attach login form listener
    document.querySelector("form").addEventListener("submit", (event) => {
        event.preventDefault();
        login();
    });

    // Attach other listeners
    jsonViewer.addEventListener("click", (e) => {
        if (e.target === jsonViewer) closeJsonViewer();
    });
    document.getElementById("file-filter").addEventListener("input", filterFiles);
    document.getElementById("content-search").addEventListener("input", () => {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(performSearch, 500);
    });
    document.getElementById("search-next").addEventListener("click", () => navigateSearch(1));
    document.getElementById("search-prev").addEventListener("click", () => navigateSearch(-1));

    // Attempt auto-login
    const user = localStorage.getItem("auth_user");
    const pass = localStorage.getItem("auth_pass");

    if (user && pass) {
        document.getElementById("user").value = user;
        document.getElementById("pass").value = pass;
        await login();
    }
});