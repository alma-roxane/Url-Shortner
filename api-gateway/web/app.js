const form = document.getElementById("shorten-form");
const result = document.getElementById("result");
const errBox = document.getElementById("error");
const submitBtn = document.getElementById("submitBtn");
const shortUrlEl = document.getElementById("shortUrl");
const shortCodeEl = document.getElementById("shortCode");
const longUrlOutEl = document.getElementById("longUrlOut");
const copyBtn = document.getElementById("copyBtn");
const copyState = document.getElementById("copyState");

let latestShortURL = "";

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  hideErr();
  copyState.textContent = "";

  const longUrl = form.longUrl.value.trim();
  const customCode = form.customCode.value.trim();
  const ttlDaysRaw = form.ttlDays.value.trim();
  const apiKey = form.apiKey.value.trim();

  if (!isValidURL(longUrl)) {
    return showErr("Please enter a valid URL including http:// or https://");
  }

  const payload = { longUrl };
  if (customCode) payload.customCode = customCode;
  if (ttlDaysRaw !== "") payload.ttlDays = Number(ttlDaysRaw);

  submitBtn.disabled = true;
  submitBtn.textContent = "Creating...";

  try {
    const res = await fetch("/api/v1/urls", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-API-Key": apiKey,
      },
      body: JSON.stringify(payload),
    });

    const data = await parseJSON(res);

    if (!res.ok) {
      throw new Error(data.error || "Unable to shorten URL");
    }

    latestShortURL = data.shortUrl;
    shortUrlEl.textContent = data.shortUrl;
    shortUrlEl.href = data.shortUrl;
    shortCodeEl.textContent = data.code;
    longUrlOutEl.textContent = data.longUrl;
    longUrlOutEl.href = data.longUrl;

    result.classList.remove("hidden");
  } catch (err) {
    showErr(err.message || "Something went wrong");
  } finally {
    submitBtn.disabled = false;
    submitBtn.textContent = "Create Short URL";
  }
});

copyBtn.addEventListener("click", async () => {
  if (!latestShortURL) return;
  try {
    await navigator.clipboard.writeText(latestShortURL);
    copyState.textContent = "Copied";
  } catch {
    copyState.textContent = "Copy failed";
  }
});

function showErr(message) {
  errBox.textContent = message;
  errBox.classList.remove("hidden");
  result.classList.add("hidden");
}

function hideErr() {
  errBox.classList.add("hidden");
  errBox.textContent = "";
}

function isValidURL(value) {
  try {
    const parsed = new URL(value);
    return parsed.protocol === "http:" || parsed.protocol === "https:";
  } catch {
    return false;
  }
}

async function parseJSON(res) {
  try {
    return await res.json();
  } catch {
    return {};
  }
}
