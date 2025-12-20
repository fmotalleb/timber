let AUTH_HEADER = null;
let tailInterval = null;

const filesContainer = document.getElementById("files");
const output = document.getElementById("output");
const jsonViewer = document.getElementById("json-viewer");
const jsonOutput = document.getElementById("json-output");

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
                const regex = new RegExp(`(${level})`, 'g');
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
    renderOutput(text);
}

async function login() {
    const user = document.getElementById("user").value;
    const pass = document.getElementById("pass").value;

    AUTH_HEADER = "Basic " + btoa(user + ":" + pass);

    try {
        await authFetch("./filesystem/ls");

        // store credentials in localStorage
        localStorage.setItem("auth_user", user);
        localStorage.setItem("auth_pass", pass);

        document.getElementById("login").classList.add("hidden");
        document.getElementById("app").classList.remove("hidden");

        loadFiles();
    } catch (e) {
        document.getElementById("login-error").textContent =
            "Authentication failed";
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

window.addEventListener("load", async () => {
    const user = localStorage.getItem("auth_user");
    const pass = localStorage.getItem("auth_pass");

    if (user && pass) {
        AUTH_HEADER = "Basic " + btoa(user + ":" + pass);
        try {
            await authFetch("./filesystem/ls");
            document.getElementById("login").classList.add("hidden");
            document.getElementById("app").classList.remove("hidden");
            loadFiles();
        } catch (e) {
            // invalid credentials, clear storage
            localStorage.removeItem("auth_user");
            localStorage.removeItem("auth_pass");
            AUTH_HEADER = null;
        }
    }
    
    // Close modal on outside click
    jsonViewer.addEventListener("click", (e) => {
        if (e.target === jsonViewer) {
            closeJsonViewer();
        }
    });
});

function createFileRow(path) {
    const row = document.createElement("div");
    row.className = "file";

    const pathEl = document.createElement("div");
    pathEl.className = "path";
    pathEl.textContent = path;

    const lines = document.createElement("input");
    lines.type = "number";
    lines.min = 1;
    lines.value = 10;
    lines.style.width = "50px";

    const cat = document.createElement("button");
    cat.textContent = "cat";
    cat.onclick = async () => {
        stopFollow();
        fetchText(`./filesystem/cat?path=${encodePath(path)}`);
    };

    const head = document.createElement("button");
    head.textContent = "head";
    head.onclick = async () => {
        stopFollow();
        fetchText(`./filesystem/head?path=${encodePath(path)}&lines=${lines.value}`);
    };

    const tail = document.createElement("button");
    tail.textContent = "tail";
    tail.onclick = async () => {
        stopFollow();
        fetchText(`./filesystem/tail?path=${encodePath(path)}&lines=${lines.value}&follow=false`);
    };

    const follow = document.createElement("button");
    follow.textContent = "follow";
    follow.onclick = async () => {
        stopFollow();
        output.innerHTML = "";

        const url = `./filesystem/tail?path=${encodePath(path)}&lines=${lines.value}&follow=true`;

        try {
            const res = await authFetch(url);
            const reader = res.body.getReader();
            const decoder = new TextDecoder();

            async function readChunk() {
                const { done, value } = await reader.read();
                if (done) return;
                const chunk = decoder.decode(value, { stream: true });
                
                const lines = chunk.split("\n");
                lines.forEach(line => {
                    if (line.trim() === "") return;
                    const lineEl = document.createElement("div");
                    lineEl.className = "log-line";
                    const logLevels = ["WARN", "FINE", "ERROR", "OK", "INFO"];
                    let hasLevel = false;
                    logLevels.forEach(level => {
                        if (line.includes(level)) {
                            const regex = new RegExp(`(${level})`, 'g');
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
                readChunk();
            }

            readChunk();
        } catch (e) {
            renderOutput("\n[Error: " + e.message + "]");
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
            a.download = path.split("/").pop(); // file name
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
        } catch (e) {
            renderOutput("\n[Error: " + e.message + "]");
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
            renderOutput("Error fetching file for JSON view.\n\n[Error: " + e.message + "]");
        }
    };


    row.append(
        pathEl,
        lines,
        cat,
        head,
        tail,
        follow,
        download,
        viewJson
    );

    return row;
}

async function loadFiles() {
    const res = await authFetch("./filesystem/ls");
    const files = await res.json();

    filesContainer.innerHTML = "";
    files.forEach(f => filesContainer.appendChild(createFileRow(f)));
}

function openJsonViewer(fileContent) {
    jsonOutput.innerHTML = ""; // Clear previous JSON output
    const lines = fileContent.split('\n').filter(line => line.trim() !== "");
    const jsonObjects = [];

    lines.forEach(line => {
        try {
            jsonObjects.push(JSON.parse(line));
        } catch (e) {
            // Not a valid JSON line, just ignore it.
        }
    });

    if (jsonObjects.length === 0) {
        jsonOutput.textContent = "No valid JSON objects found per line in this file.";
        jsonViewer.classList.remove("hidden");
        return;
    }

    // Create a table
    const table = document.createElement('table');
    table.className = 'json-table';

    // Create table header
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

    // Create table body
    const tbody = document.createElement('tbody');
    jsonObjects.forEach(obj => {
        const row = document.createElement('tr');
        headers.forEach(header => {
            const cell = document.createElement('td');
            const value = obj[header];
            if (typeof value === 'object' && value !== null) {
                cell.textContent = JSON.stringify(value, null, 2);
            } else {
                cell.textContent = value;
            }
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