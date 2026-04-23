import fs from "fs";

const API_KEYS = [
  "7d85ddac-5690-4453-aa58-751ae68ab060",
  "b75c60e5-1ea5-4f5e-8510-58e6a728f249",
  "aefd2992-ea2f-4b0e-aae4-2009e29cbf6a",
];

let keyIndex = 0;
function getNextKey() {
  return API_KEYS[keyIndex++ % API_KEYS.length];
}

const CONCURRENCY_LIMIT = 3; // Reduce from 5 → 3 to be safer
const RETRY_LIMIT = 3;
const RETRY_DELAY_MS = 1000;

const sleep = (ms) => new Promise((res) => setTimeout(res, ms));

// Step 1: Extract IDs with team association
function extractIds(filePath) {
  const content = fs.readFileSync(filePath, "utf-8");
  const uuidRegex =
    /[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}/g;
  const result = [];
  const seen = new Set();

  const sections = content.split(/<h4>/);
  for (const section of sections.slice(1)) {
    const teamMatch = section.match(/^([^<]+)<\/h4>/);
    if (!teamMatch) continue;
    const team = teamMatch[1].trim();

    for (const id of section.match(uuidRegex) || []) {
      if (!seen.has(id)) {
        seen.add(id);
        result.push({ id, team });
      }
    }
  }

  return result;
}

// Step 2: Fetch player with retry
async function fetchPlayer({ id, team }, attempt = 1) {
  const apiKey = getNextKey();
  const url = `https://api.cricapi.com/v1/players_info?apikey=${apiKey}&id=${id}`;

  try {
    const res = await fetch(url);

    // Handle non-OK HTTP responses (e.g. 429 rate limit)
    if (!res.ok) {
      throw new Error(`HTTP ${res.status}`);
    }

    const json = await res.json();
    if (!json?.data) return null;

    const { name, dateOfBirth, role, battingStyle, bowlingStyle, country, playerImg } = json.data;
    return { id, name, dateOfBirth, role, battingStyle, bowlingStyle, country, playerImg, team };

  } catch (err) {
    const isRetryable =
      err.cause?.code === "ENOTFOUND" ||
      err.cause?.code === "ECONNRESET" ||
      err.message?.includes("HTTP 429") ||
      err.message?.includes("fetch failed");

    if (isRetryable && attempt <= RETRY_LIMIT) {
      console.warn(`⚠️  Retry ${attempt}/${RETRY_LIMIT} for ${id} (${err.message})`);
      await sleep(RETRY_DELAY_MS * attempt); // exponential-ish backoff
      return fetchPlayer({ id, team }, attempt + 1);
    }

    console.error(`❌ Failed permanently for ${id}:`, err.message);
    return null;
  }
}

// Step 3: Concurrency controller
async function runWithConcurrency(ids, limit) {
  const results = [];
  let index = 0;

  async function worker() {
    while (index < ids.length) {
      const item = ids[index++];
      const result = await fetchPlayer(item);
      if (result) {
        results.push(result);
        console.log(`✅ ${result.name} (${result.team})`);
      }
      await sleep(200); // small breathing room between requests
    }
  }

  await Promise.all(Array.from({ length: limit }, () => worker()));
  return results;
}

// Step 4: Convert to CSV
function convertToCSV(data) {
  const headers = ["id", "name", "dateOfBirth", "role", "battingStyle", "bowlingStyle", "country", "playerImg", "team"];
  const rows = data.map((obj) =>
    headers.map((field) => `"${String(obj[field] ?? "").replace(/"/g, '""')}"`).join(",")
  );
  return [headers.join(","), ...rows].join("\n");
}

// Step 5: Main
async function main() {
  const ids = extractIds("all.html");
  console.log(`🔍 Found ${ids.length} players across ${new Set(ids.map((p) => p.team)).size} teams`);

  const players = await runWithConcurrency(ids, CONCURRENCY_LIMIT);

  const csv = convertToCSV(players);
  fs.writeFileSync("players.csv", csv, "utf-8");
  console.log(`✅ Saved ${players.length} players to players.csv`);
}

main();