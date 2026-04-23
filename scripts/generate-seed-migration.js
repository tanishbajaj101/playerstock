import fs from 'fs';

const TEAM_LOGOS = {
  'Chennai Super Kings':         '/logos/Chennai Super Kings.webp',
  'Delhi Capitals':              '/logos/Delhi Capitals.webp',
  'Gujarat Titans':              '/logos/Gujarat Titans.webp',
  'Kolkata Knight Riders':       '/logos/Kolkata Knight Riders.webp',
  'Lucknow Super Giants':        '/logos/Lucknow Super Giants.webp',
  'Mumbai Indians':              '/logos/Mumbai Indians.webp',
  'Punjab Kings':                '/logos/Punjab Kings.webp',
  'Rajasthan Royals':            '/logos/Rajasthan Royals.png',
  'Royal Challengers Bengaluru': '/logos/Royal Challengers Bangalore.webp',
  'Sunrisers Hyderabad':         '/logos/Sunrisers Hyderabad.webp',
};

function parseCSVLine(line) {
  const fields = [];
  let i = 0;
  while (i < line.length) {
    if (line[i] === '"') {
      let val = '';
      i++;
      while (i < line.length) {
        if (line[i] === '"' && line[i + 1] === '"') { val += '"'; i += 2; }
        else if (line[i] === '"') { i++; break; }
        else { val += line[i++]; }
      }
      if (i < line.length && line[i] === ',') i++;
      fields.push(val);
    } else {
      let val = '';
      while (i < line.length && line[i] !== ',') val += line[i++];
      if (i < line.length && line[i] === ',') i++;
      fields.push(val);
    }
  }
  return fields;
}

function parseCSV(text) {
  const lines = text.split('\n').filter(l => l.trim());
  const headers = parseCSVLine(lines[0]);
  return lines.slice(1).map(line => {
    const vals = parseCSVLine(line);
    const obj = {};
    headers.forEach((h, idx) => { obj[h] = vals[idx] ?? ''; });
    return obj;
  });
}

function q(s) {
  if (!s || s === '--') return 'NULL';
  return `'${s.replace(/'/g, "''")}'`;
}

function qDate(s) {
  if (!s) return 'NULL';
  const d = s.split('T')[0];
  if (!d || d.length < 10) return 'NULL';
  return `'${d}'`;
}

function makeSymbol(name, seen) {
  let base = name.toUpperCase().replace(/[^A-Z0-9]+/g, '_').replace(/^_+|_+$/g, '');
  if (base.length > 30) base = base.slice(0, 30).replace(/_+$/, '');
  let sym = base;
  let n = 2;
  while (seen.has(sym)) { sym = `${base}_${n++}`; }
  seen.add(sym);
  return sym;
}

const csvContent = fs.readFileSync('players.csv', 'utf-8');
const players = parseCSV(csvContent);

const seen = new Set();
const rows = players.map(p => {
  const id = p.id.toLowerCase();
  const symbol = makeSymbol(p.name, seen);
  const teamLogo = TEAM_LOGOS[p.team] ? q(TEAM_LOGOS[p.team]) : 'NULL';

  return `('${id}', ${q(symbol)}, ${q(p.name)}, '', ${q(p.country)}, ${q(p.role)}, ${qDate(p.dateOfBirth)}, ${q(p.battingStyle)}, ${q(p.bowlingStyle)}, ${q(p.playerImg)}, 0, ${q(p.team)}, ${teamLogo})`;
});

const sql = `-- +migrate Up
ALTER TABLE assets ADD COLUMN IF NOT EXISTS team      TEXT NOT NULL DEFAULT '';
ALTER TABLE assets ADD COLUMN IF NOT EXISTS team_logo TEXT;

-- Clear all previous assets (CASCADE wipes positions, orders, trades, special_coin_uses, price_snapshots)
TRUNCATE assets CASCADE;

INSERT INTO assets (id, symbol, name, description, nationality, role, date_of_birth, batting_style, bowling_style, player_img, supply_used, team, team_logo) VALUES
${rows.join(',\n')};
`;

const outPath = 'backend/internal/db/migrations/000008_add_team.up.sql';
fs.writeFileSync(outPath, sql, 'utf-8');
console.log(`Generated ${outPath} with ${players.length} players`);
