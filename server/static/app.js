
let AUTH_HEADER = null;
let tailInterval = null;

const filesContainer = document.getElementById("files");
const output = document.getElementById("output");

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

async function fetchText(url) {
    const res = await authFetch(url);
    return res.text();
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
    [...filesContainer.children].forEach((a) => filesContainer.removeChild(a));
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

    const cat = document.createElement("button");
    cat.textContent = "cat";
    cat.onclick = async () => {
        stopFollow();
        output.textContent = await fetchText(
            `./filesystem/cat?path=${encodePath(path)}`
        );
    };

    const head = document.createElement("button");
    head.textContent = "head";
    head.onclick = async () => {
        stopFollow();
        output.textContent = await fetchText(
            `./filesystem/head?path=${encodePath(path)}&lines=${lines.value}`
        );
    };

    const tail = document.createElement("button");
    tail.textContent = "tail";
    tail.onclick = async () => {
        stopFollow();
        output.textContent = await fetchText(
            `./filesystem/tail?path=${encodePath(path)}&lines=${lines.value}&follow=false`
        );
    };

    const follow = document.createElement("button");
    follow.textContent = "follow";
    follow.onclick = async () => {
        stopFollow();
        output.textContent = "";

        const url =
            `./filesystem/tail?path=${encodePath(path)}` +
            `&lines=${lines.value}&follow=true`;

        try {
            const res = await authFetch(url);
            const reader = res.body.getReader();
            const decoder = new TextDecoder();

            async function readChunk() {
                const { done, value } = await reader.read();
                if (done) return;
                output.textContent += decoder.decode(value, { stream: true });
                output.scrollTop = output.scrollHeight;
                readChunk();
            }

            readChunk();
        } catch (e) {
            output.textContent += "\n[Error: " + e.message + "]";
        }
    };

    row.append(
        pathEl,
        lines,
        cat,
        head,
        tail,
        follow
    );

    return row;
}

async function loadFiles() {
    const res = await authFetch("./filesystem/ls");
    const files = await res.json();

    filesContainer.innerHTML = "";
    files.forEach(f => filesContainer.appendChild(createFileRow(f)));
}
