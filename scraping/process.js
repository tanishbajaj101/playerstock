import fs from "fs";

const API_KEYS = [
  "18ab2b1f-375b-4571-a9e7-b1eff20a40c9",
  "YOUR_SECOND_API_KEY",
  "YOUR_THIRD_API_KEY",
];

let keyIndex = 0;
function getNextKey() {
  const key = API_KEYS[keyIndex % API_KEYS.length];
  keyIndex++;
  return key;
}

const CONCURRENCY_LIMIT = 5;

// Step 1: Extract IDs with team association
function extractIds(filePath) {
  const content = fs.readFileSync(filePath, "utf-8");
  const uuidRegex = /[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}/g;
  const result = [];
  const seen = new Set();

  const sections = content.split(/<h4>/);
  for (const section of sections.slice(1)) {
    const teamMatch = section.match(/^([^<]+)<\/h4>/);
    if (!teamMatch) continue;
    const team = teamMatch[1].trim();

    for (const id of (section.match(uuidRegex) || [])) {
      if (!seen.has(id)) {
        seen.add(id);
        result.push({ id, team });
      }
    }
  }

  return result;
}

// Step 2: Fetch player
async function fetchPlayer({ id, team }) {
  const apiKey = getNextKey();
  const url = `https://api.cricapi.com/v1/players_info?apikey=${apiKey}&id=${id}`;

  try {
    const res = await fetch(url);
    const json = await res.json();

    if (!json?.data) return null;

    const {
      name,
      dateOfBirth,
      role,
      battingStyle,
      bowlingStyle,
      country,
      playerImg,
    } = json.data;

    return { id, name, dateOfBirth, role, battingStyle, bowlingStyle, country, playerImg, team };

  } catch (err) {
    console.error(`❌ Error for ${id}`, err);
    return null;
  }
}

// Step 3: Concurrency controller
async function runWithConcurrency(ids, limit) {
  const results = [];
  let index = 0;

  async function worker() {
    while (index < ids.length) {
      const currentIndex = index++;
      const item = ids[currentIndex];

      const result = await fetchPlayer(item);
      if (result) {
        results.push(result);
        console.log(`✅ ${result.name} (${result.team})`);
      }
    }
  }

  const workers = Array.from({ length: limit }, () => worker());
  await Promise.all(workers);

  return results;
}

// Step 4: Convert to CSV
function convertToCSV(data) {
  const headers = [
    "id", "name", "dateOfBirth", "role",
    "battingStyle", "bowlingStyle", "country", "playerImg", "team",
  ];

  const rows = data.map(obj =>
    headers.map(field => {
      const value = obj[field] ?? "";
      return `"${String(value).replace(/"/g, '""')}"`;
    }).join(",")
  );

  return [headers.join(","), ...rows].join("\n");
}

// Step 5: Main
async function main() {
  const ids = extractIds("all.html");

  console.log(`🔍 Found ${ids.length} players across ${new Set(ids.map(p => p.team)).size} teams`);

  const players = await runWithConcurrency(ids, CONCURRENCY_LIMIT);

  console.log("\n🎯 Writing CSV...");

  const csv = convertToCSV(players);
  fs.writeFileSync("players.csv", csv, "utf-8");

  console.log("✅ Saved to players.csv");
}

main();