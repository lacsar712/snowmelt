const API = "/api";
const trendData = [];
const MAX_POINTS = 60;

async function fetchJSON(path, opts) {
  const res = await fetch(API + path, opts);
  const body = await res.json();
  if (!body.ok) throw new Error(body.error || "request failed");
  return body.data;
}

function setMessage(text, kind) {
  const el = document.getElementById("message");
  el.textContent = text;
  el.className = "message" + (kind ? " " + kind : "");
}

function renderSnapshot(snap) {
  document.getElementById("unit-label").textContent = "Unit: " + snap.unit_id;
  document.getElementById("state").textContent = snap.state;
  document.getElementById("pressure").textContent = snap.runway.steam_pressure_psi.toFixed(0);
  document.getElementById("mw").textContent = snap.runway.output_mw.toFixed(1);
  document.getElementById("glycol-level").textContent = snap.glycol.level_percent.toFixed(1);
  document.getElementById("burner").textContent = snap.heatbank.burner_phase;
  document.getElementById("o2").textContent = snap.heatbank.excess_o2_pct.toFixed(2);
  trendData.push(snap.runway.steam_pressure_psi);
  if (trendData.length > MAX_POINTS) trendData.shift();
  drawTrend();
}

async function refresh() {
  try {
    const snap = await fetchJSON("/snapshot");
    renderSnapshot(snap);
    const health = await fetchJSON("/health/plant");
    const warmup = await fetchJSON("/warmup");
    document.getElementById("health").textContent =
      JSON.stringify({ health, warmup }, null, 2);
  } catch (err) {
    setMessage(err.message, "error");
  }
}

async function postAction(path) {
  try {
    setMessage("Sending…");
    const snap = await fetchJSON(path, { method: "POST" });
    renderSnapshot(snap);
    setMessage("OK", "ok");
  } catch (err) {
    setMessage(err.message, "error");
  }
}

function drawTrend() {
  const canvas = document.getElementById("trend");
  const ctx = canvas.getContext("2d");
  const w = canvas.width;
  const h = canvas.height;
  ctx.clearRect(0, 0, w, h);
  if (trendData.length < 2) return;
  const max = Math.max(...trendData, 1);
  ctx.strokeStyle = "#3d9be9";
  ctx.lineWidth = 2;
  ctx.beginPath();
  trendData.forEach((v, i) => {
    const x = (i / (MAX_POINTS - 1)) * w;
    const y = h - (v / max) * (h - 10) - 5;
    if (i === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  });
  ctx.stroke();
}

document.getElementById("btn-preheat").addEventListener("click", () => postAction("/preheat/start?holder=hmi"));
document.getElementById("btn-preheat-done").addEventListener("click", () => postAction("/preheat/complete?holder=hmi"));
document.getElementById("btn-ignite").addEventListener("click", () => postAction("/ignite?holder=hmi"));
document.getElementById("btn-reset").addEventListener("click", () => postAction("/trip/reset?holder=hmi"));

refresh();
setInterval(refresh, 3000);
